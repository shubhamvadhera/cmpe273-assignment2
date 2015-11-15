# CRUD Location Service

This location service has the following REST endpoints to store and retrieve locations. All the data is persisted into MongoLab.
Google Map Api has been used to lookup coordinates of a location.

###1. POST - Create New Location ( "/locations" )
Request:
```http
POST /locations
```
```json
{
   "name" : "John Smith",
   "address" : "123 Main St",
   "city" : "San Francisco",
   "state" : "CA",
   "zip" : "94113"
}
```
Response:
```http
HTTP Response Code: 201
```
```json
{
   "id" : 589d1a94f0gh201275fbe169,
   "name" : "John Smith",
   "address" : "123 Main St",
   "city" : "San Francisco",
   "state" : "CA",
   "zip" : "94113",
   "coordinate" : { 
      "lat" : 38.4220352,
     "lng" : -222.0841244
   }
}
```
###2. GET - Get a saved Location ( "/locations/{location_id}" )
Request:
```http
GET /locations/589d1a94f0gh201275fbe169
```

Response:
```http
HTTP Response Code: 200
```
```json
{
   "id" : 589d1a94f0gh201275fbe169,
   "name" : "John Smith",
   "address" : "123 Main St",
   "city" : "San Francisco",
   "state" : "CA",
   "zip" : "94113",
   "coordinate" : { 
      "lat" : 38.4220352,
     "lng" : -222.0841244
   }
}
```
###PUT - Update a Location ( "/locations/{location_id}" )
Request:
```http
PUT /locations/589d1a94f0gh201275fbe169
```
Response:
```http
HTTP Response Code: 202
```
```json
{
   "id" : 589d1a94f0gh201275fbe169,
   "name" : "John Smith",
   "address" : "1600 Amphitheatre Parkway",
   "city" : "Mountain View",
   "state" : "CA",
   "zip" : "94043"
   "coordinate" : { 
      "lat" : 37.4220352,
     "lng" : -122.0841244
   }
}
```
##4. DELETE - Delete a saved Location ( "DELETE /locations/{location_id}" )
Request:
```http
DELETE  /locations/589d1a94f0gh201275fbe169
```
Response:
```http
HTTP Response Code: 200
```
