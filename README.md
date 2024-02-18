## About CalculationServer
CalculationServer is a server that calculates expressions within a specified operation calculation time.

The project also utilizes a PostgresSQL database, [Bootstrap](https://getbootstrap.com), and [HTMX](https://htmx.org). Bootstrap is employed to create a fast and visually appealing interface, while HTMX is used to send requests and handle them in a convenient manner.

## Getting started
To run the server, you need:
- [Go](https://golang.org/dl) version 1.22 or later.
- [PostgreSQL](https://postgresql.org/download)

To start working with the server, follow these steps:
1. Clone the repository:
   ```bash
   git clone https://github.com/LLIEPJIOK/CalculatingServer.git
   ```
2. Open the `expressions/database.go` file and change the `port` and `password` to match your PostgreSQL port and password.
3. Type the following command in the console:
   ```bash
   go run *.go
   ```
4. Open [`localhost:8080`](http://localhost:8080) in your browser.

## Code structure
1. `main.go` - file for the server where requests are processed.
2. `expressions/expressions.go` - file for handling expressions.
3. `expressions/database.go` - file for interacting with the PostgreSQL database.
4. `static/` - folder for the visual interface. It contains CSS styles, JS script for clearing the input field after sending an expression, and HTML templates with Bootstrap and HTMX.

## How it works
The program consists of a server that handles all requests and agents, which are goroutines responsible for calculating expressions asynchronously.

When you send a request (e.g., open the page, submit an expression, etc.), the server handles it, and there are several possible scenarios:

1. **Request to calculate an expression:**
   
   The server first parses the expression to ensure its validity.
   - If the expression is valid, the server conveys it to the agents through a channel. At some point, one agent picks it up, calculates the expression, and then updates the result.
   - If the expression is invalid, the server sets a parsing error in the expression status.

2. **Request to update the operation calculation time:**

   The server updates this data in the database.

3. **Other requests:**

   The server retrieves information from the database and displays it.

*Note: Currently, there is no automatic updating of data on the page; to see the changes, you must reload the page.*

## Example of usage
![Example of usage](https://github.com/LLIEPJIOK/CalculatingServer/tree/master/images/ServerUsage.gif)