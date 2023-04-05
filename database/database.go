package database

import (
	"context"
	"fmt"
	"log"

	dbDriver "github.com/arangodb/go-driver"
	arangoHttp "github.com/arangodb/go-driver/http"
	"github.com/spf13/viper"
)

type Airport struct {
	Key     string `json:"_key"`
	Name    string `json:"name"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}

type Flight struct {
	FlightNum     int     `json:"FlightNum"`
	From          Airport `json:"From"`
	To            Airport `json:"To"`
	Departure     string  `json:"DepTimeUTC"`
	Arrival       string  `json:"ArrTimeUTC"`
	Duration      float32 `json:"Duration"`
	DurFormatted  string  `json:"DurFormatted"`
	UniqueCarrier string  `json:"UniqueCarrier"`
	TailNum       string  `json:"TailNum"`
}

type Connection struct {
	Flights       []Flight `json:"Flights"`
	TotalDuration int      `json:"TotalDuration"`
	TotFormatted  string   `json:"TotFormatted"`
	FlightsCount  int      `json:"FlightsCount"`
}

var (
	db  dbDriver.Database
	ctx context.Context
)

func Setup() {
	var (
		err    error
		client dbDriver.Client
		conn   dbDriver.Connection
	)

	conn, err = arangoHttp.NewConnection(arangoHttp.ConnectionConfig{
		Endpoints: []string{fmt.Sprintf("%v", getConf("DB_HOST"))},
	})
	if err != nil {
		log.Panicf("Failed to create HTTP connection: %v", err)
	}

	client, err = dbDriver.NewClient(dbDriver.ClientConfig{
		Connection: conn,
		Authentication: dbDriver.BasicAuthentication(
			getConf("DB_USER"),
			getConf("DB_PASSWORD"),
		),
	})
	if err != nil {
		log.Panicf("Failed to create database client: %v", err)
	}
	ctx = context.Background()

	db, err = client.Database(ctx, getConf("DB_NAME"))
	if err != nil {
		log.Panicf("Failed to open existing database: %v", err)
	}
}

func FindConnections(date string, from string, to string, limit int) ([]any, error) {
	var (
		/*
			Previously had typed variables but when using these, the append updated flights for all conneciton to be the same as the last one
			connection  Connection
			connections []Connection
		*/
		connection  interface{}
		connections []any
	)

	query := `
		WITH airports
		LET y = DATE_YEAR(@date)
		LET m = DATE_MONTH(@date)
		LET d = DATE_DAY(@date)
		
		FOR v, e, p IN 2 /*2..3 - for 3-flight connections*/ OUTBOUND @from flights
		OPTIONS {
			order: "bfs",
			uniqueVertices: 'path',
			uniqueEdges: 'path'
		}
		
		FILTER
			v._id == @to
			AND p.edges[*].Year ALL == y
			AND p.edges[*].Month ALL == m
			AND p.edges[*].Day ALL == d
			AND DATE_ADD(p.edges[0].ArrTimeUTC, 20, 'minutes') < p.edges[1].DepTimeUTC
			//AND DATE_ADD(p.edges[1].ArrTimeUTC, 20, 'minutes') < p.edges[2].DepTimeUTC -- uncomment for 3-flight connections
		LET TotalDuration = DATE_DIFF(FIRST(p.edges).DepTimeUTC, LAST(p.edges).ArrTimeUTC, 'i')
		SORT TotalDuration ASC
		LIMIT @limit
		LET TotFormatted = CONCAT(FLOOR(TotalDuration / 60), 'h:', TotalDuration % 60, 'min')
		LET Flights = (
			FOR flight IN p.edges
			LET Duration = DATE_DIFF(flight.DepTimeUTC, flight.ArrTimeUTC, 'i', true)
			RETURN {
				FlightNum: flight.FlightNum, DepTime: flight.DepTime, ArrTime: flight.ArrTime,
				DepTimeUTC: DATE_FORMAT(flight.DepTimeUTC, "%yyyy-%mm-%dd %hh:%ii:%ss"),
				ArrTimeUTC: DATE_FORMAT(flight.ArrTimeUTC, "%yyyy-%mm-%dd %hh:%ii:%ss"),
				Duration: Duration,
				DurFormatted: CONCAT(FLOOR(Duration / 60), 'h:', Duration % 60, 'min'),
				TailNum: flight.TailNum,
				UniqueCarrier: flight.UniqueCarrier,
				From: DOCUMENT(flight._from),
				To: DOCUMENT(flight._to)
			}
		)
		RETURN { Flights, TotalDuration, TotFormatted, FlightsCount: LENGTH(Flights) }`

	vars := map[string]interface{}{
		"from":  "airports/" + from,
		"to":    "airports/" + to,
		"date":  date,
		"limit": limit,
	}

	var cursor, err = db.Query(ctx, query, vars)
	if err != nil {
		log.Panicf("Query failed: %v", err)
		return nil, err
	}

	defer cursor.Close()

	for {
		_, err = cursor.ReadDocument(ctx, &connection)

		if dbDriver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Panicf("Doc returned: %v", err)
			return nil, err
		} else {
			connections = append(connections, connection)
		}
	}
	fmt.Printf("%+v\n", connections)
	return connections, nil
}

func GetAirports() ([]Airport, error) {
	var (
		airport  Airport
		airports []Airport
	)

	query := `
		FOR a IN airports
		SORT a.name
		RETURN a`

	var cursor, err = db.Query(ctx, query, nil)
	if err != nil {
		log.Panicf("Query failed: %v", err)
		return nil, err
	}

	defer cursor.Close()

	for {
		_, err = cursor.ReadDocument(ctx, &airport)

		if dbDriver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Panicf("Doc returned: %v", err)
			return nil, err
		} else {
			airports = append(airports, airport)
		}
	}
	return airports, nil
}

func getConf(name string) string {
	return fmt.Sprintf("%v", viper.Get(name))
}
