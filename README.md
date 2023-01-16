# 📋 ToDo App - a databases project for school


<p>A simple todo app API with jwt user authentication, written with golang and the fiber framework, using the go-sql-driver as an interface to a MySQL database.</p>

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

## API routes - once you have started the server, you can start interacting with it by sending and recieving JSON data, a good tool for testing such APIs is [postman](https://www.postman.com)

- `localhost:8000/api/register` - register a new user
  -  JSON in the following format should be provided:
  ```
  {
    "username": "<username>",
    "email": "<email>",
    "password": "<password>"  
  }
  ```

- `localhost:8000/api/login` - login
  -  JSON in the following format should be provided:
  ```
  {
    "email": "<email>",
    "password": "<password>"
  } 
  ```

- `localhost:8000/api/user` - get logged in user
  - No JSON needs to be provided

- `localhost:8000/api/user/update` - update user with new credentials
  - JSON in the following format should be provided: 
  ```
  {
    "id":<id>, // integer
    "username": "<username>",
    "email": "<email>",
    "password": "<password>"
  }
  ```
  
- `localhost:8000/api/activity/new` - create new activity
  - JSON in the following format should be provided:
  ```
  {
    "title": "<title>",
    "body": "<body>",
    "due": "<due>", // should be in format "2006-12-11T12:55:23+02:00"
    "group_id": <group_id> //integer or null
  }
  ```
  
- `localhost:8000/api/activity/update` - update activity
  - JSON in the following format should be provided:
  ```
  {
    "id": <id>, // integer
    "title": "<title>",
    "body": "body",
    "due": "<due>" // // should be in format "2006-12-11T12:55:23+02:00"
  }
  ```
  
- `localhost:8000/api/activity/<id>/delete` <id> is the integer value of the activity id - delete an activity
  - No JSON needs to be provided
  
- `localhost:8000/api/activity/group` - add an activity to a group
  - JSON in the following format should be provided:
  ```
  {
    "activity_id": <id>, // integer
    "group_id": <group_id> // integer
  }
  ```
  
- `localhost:8000/api/activities?<key>=<value>` - get 10 activities and their groups, `A JOIN CLOSED IS USED HERE`
  - The following URL parameters must be provided:
    - `start` - as the data is paginated, `start` determines which set of 10 activities to get from the database, default value is `1`
    - `desc` - if `desc=true`, the most recent posts will be shown first, if `desc=false`, the oldest posts are going to be first
    - `all` - determines that nothing specific is being searched for, gets 10 activities from all available
    - `search` - filters the activities by the value provided, activities containing data similar to the searched in their titles and bodies are returned
    !NOTE - if both `all` and `search` are provided, the data returned is not going to be filtered by the search
    
- `localhost:8000/api/group/new` - create a new group
  - JSON in the following format should be provided:
  ```
  {
    "name": "<name>" // provide the name of the group
  }
  ```
  
- `localhost:8000/api/group/<id>/delete` - <id> is an integer value that indicates the group that is to be deleted

- `localhost:8000/api/group/update` - updates a group
  - JSON in the following format should be provided:
  ```
  {
    "id": <group_id>, // id of the group to be updated
    "name": "<name>" // new name of the group
  }
  ```
  
- `localhost:8000/api/group/<id>/get` - <id> is an integer value that indicates the group that is to be returned in the form of JSON; returns the group details and all activities related to it

- `localhost:8000/api/group/<id>/delete` - <id> is an integer value that indicates the group that is to be deleted; deletes a group

- `localhost:8000/api/groups` - returns all groups and their activities
