## About calculation server

Read this in other languages: [Русский](https://github.com/LLIEPJIOK/CalculatingServer/blob/master/README.ru.md).

Calculation server is a server that calculates expressions within a specified operation calculation time.

The project uses a [PostgresSQL](https://www.postgresql.org) database, [Bootstrap](https://getbootstrap.com) and [HTMX](https://htmx.org). Bootstrap is used to create a fast and visually appealing interface, while HTMX is used to send requests and handle them in a convenient manner.

## Getting started
To run the server, you only need [Docker](https://www.docker.com/products/docker-desktop/) running on your computer.

To start working with the server, follow these steps in console:
1. Clone the repository:
   ```bash
   git clone https://github.com/LLIEPJIOK/calculating-server.git
   ```
2. Go to the project folder:
   ```bash
   cd calculating-server
   ```
3. Start a docker container with the project:
   ```bash
   docker-compose up
   ```
4. Open [`localhost:8080`](http://localhost:8080) in your browser.

If you launched the application for the first time, you will need to register. Otherwise just log in. After that, the server gives you access to perform operations on expressions and remember you for 1 day. Now perform desired operations and see the results.

*Note: When the container starts, all tests are executed, so if one test fails, the entire program will not start.*

## Code structure
1. `main.go` - file to initialize the project.
2. `internal/controllers` - folder with files for the server where requests are processed.
3. `internal/database` - folder with files for interacting with the PostgreSQL database.
4. `internal/expression` - folder with files for processing expressions.
5. `internal/user` - folder with files for processing users.
6. `internal/workers` - folder with files for processing workers.
7. `static/` - folder for visual interface. It contains CSS styles, JS scripts, icons and HTML templates with Bootstrap and HTMX.

## How it works
The program consists of a server that handles all requests and agents, which are goroutines responsible for calculating expressions asynchronously. Goroutines also update their status (e.g., waiting, calculating, etc.) every 5 seconds if they are waiting for an expression, as well as before and after performing calculations.

Before every user request (e.g., open the page, submit an expression, etc.) the server checks for authorization. If the user is authorized then the server give him access to perform operations on expressions, but not to log in or register. Otherwise, it only allows to log in or register. Authentication was implemented using [JWT](https://jwt.io/).

After that the server handles request, and there are several possible scenarios:

1. **Request to calculate an expression:**
   
   The server add expression to database, then parses it to ensure its validity.
   - If the expression is valid, the server conveys it to the agents through a channel. At some point, one agent picks it up, calculates the expression, and then updates the result.
   - If the expression is invalid, the server sets a parsing error in the expression status.
   
   After that the server update last expressions.

2. **Request to update the operation calculation time:**

   The server updates this data in the database.

3. **Log in or register**

   The server checks the validity of the data. If the data is correct, then the server gives you access to perform operations on expressions and record your data in cookie for 1 day to remember that you have registered, otherwise it shows errors.

3. **Other requests:**

   The server retrieves information from the database and displays it.

![Working scheme](https://github.com/LLIEPJIOK/CalculatingServer/blob/master/images/WorkingScheme.png)

*Note: Currently, there is no automatic updating of data on the page. To see the changes, you must reload the page.*