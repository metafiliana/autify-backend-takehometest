# autify backend takehometest - Metafiliana

## How to run

### With docker
To run with docker we just need to build docker file image, 
```
docker build --tag autify-be-tht-metafiliana .
```

Run the container, and access container CLI
```
docker run -it autify-be-tht-metafiliana
```

Then execute the executable program
```
./fetch --metadata https://www.google.com https://autify.com
```

### Without docker
Clone this repository to your go project directory

Download go mod dependencies
```
go get && go mod tidy
```

Build executable program
```
go build -o fetch
```

Then execute the executable program
```
./fetch --metadata https://www.google.com https://autify.com
```

## Downloaded websites and assets location
* downloaded html file location are on the root folder (same as `main.go`). 
* the assets file are downloaded under `assets/{websiteName}/`. 

### Notes
`websiteName` are the url arguments appended when executing program, but the character `/` are changed to `-` to prevent character conflict with foldering structure. also the length of the `websiteName` are trimmed to max 30 character to prevent max path limit issue.