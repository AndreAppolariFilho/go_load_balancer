# Instructions
To run just create a .env file inside of each server folder, forwarded_server_1, forwarded_server_2, load_balancer with the following parameters

PORT=XXXX

This is the port that you want to build.

## First run the servers
Go to the forwarded_server_1 and run the following commands

go build

./forwarded_server_1


In another terminal go to the forwarded_server_2 and run the following commands

go build

./forwarded_server_2

In another terminal go to the load_balancer and run the following commands

go build

./load_balancer


after that just use your browser to get the infomation from this endpoint

localhost:8000/v1/