# 📋 ToDo App


<p>A simple todo app with jwt user authentication, written with golang and the fiber framework</p>

## Prerequisites
- Install and configure MySQL server
- Install the go language compiler from [here](https://go.dev/dl)

## To run the app follow the steps listed below
- Open the project directory and run `go mod tidy` to install all the dependencies
- Create a database for the app in the MySQL server - `CREATE DATABASE <DB_NAME>;`
- Create a .env file in the root of the project, which should specify:
  - `DB_USER` - a user that is registered in the MySQL server
  - `DB_PASS` - the password of the user, whose name is provided in `DB_USER`
  - `DB_NAME` - the name of the database that you created in the second step
- To run the app run `go run main.go`

## API routes
- `localhost:8000/api/register` - register a new user, JSON in the following format should be provided:
`
{
  "username": "<username>",
  "email": "<email>",
  "password": "<password>"
}
`
  - 

- `localhost:8000/api/login`
