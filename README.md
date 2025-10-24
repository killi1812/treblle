
# Treblle hackaton app


## Dependencies
taskfile
`go install github.com/go-task/task/v3/cmd/task@latest`

## Running the project

run `task setup-env` to setup enviroment for dev

`task dev` to start development database and app
`task run` to only start the app, *means you know how to run your own database*

or

download released docker images load them with `docker load -i [docer image].tar` and run `docker-compose -f deployment.yaml up` to deploy the whole stack
