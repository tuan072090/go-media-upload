# Media upload service by GO

## How to run
- Clone the repo
- Install libraries
- Run project ``PORT=8000 go run main.go``

## List route
````
GET    /files                 --> ./services/query_file.go
POST   /upload                --> ./services/upload_file.go
````
## Config environment variables

``./share/config.go``
#### Example
- PORT=8000
- MEDIA_URL=http://localhost
- MAX_UPLOAD_SIZE=10485760  ```// bytes```

## Upload an example image
Open postman, set method type to POST.

Then select Body -> form-data -> Enter "**file**" as parameter name

On right side next to value column, there will be dropdown "text, file", select File. choose your image file and post it.
