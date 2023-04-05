# This is an example REST API written in Go to implement a solution for finding best flight connections between airports

## Database
For setting up the ArangoDB database, followed the instructions provided in the manual: https://www.arangodb.com/docs/stable/deployment-single-instance-using-the-starter.html.
The dataset is provided in the folder data, containing files airports.csv and flights.csv

Import the csv files into DB using the arangoimport tool: 
```
arangoimport --file "airports.csv" --type csv --collection "airports" --create-collection 
arangoimport --file "flights.csv" --type csv --collection "flights" --create-collection --create-collection-type=edge 
```

In addition to the system defined indexes on _key, _from and _to, create an index on flights fields _to, Year, Month and Day to improve the performance of the query:
```
db.flights.ensureIndex({ type: "persistent", fields: [ "_from", "Year", "Month", "Day" ] }) 
```

## Back-end
For running the API, just run the following command in the root folder of the project
```
go run flights.go
```