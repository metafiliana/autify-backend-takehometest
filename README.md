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