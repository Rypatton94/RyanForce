# RyanForce CRM

[![Go CLI Tests](https://github.com/Rypatton94/RyanForce/actions/workflows/go-test.yml/badge.svg)](https://github.com/Rypatton94/RyanForce/actions/workflows/go-test.yml)

---

## About

RyanForce is a managed IT CRM written in Go. You can manage clients, technicians, admins, support tickets, and comments. You can use it through a CLI or a simple WebUI. There's also a real API underneath for future automation work.

Runs locally using SQLite. Tests run against an in-memory database.

---

## Features

- CLI or WebUI login with JWT session handling
- Role-based access (admin, tech, client)
- Create and assign tickets
- Comment on tickets
- Export tickets to CSV
- System logs important events
- Basic reports (ticket status, overdue tickets)

---

## CLI Commands

- login, logout
- register users (admin only)
- view, update, comment on tickets
- assign techs to tickets (admin)
- view users/accounts (admin)
- run ticket reports
- reset your password

Logs everything for auditing. Sessions expire after 24 hours.

---

## WebUI Features

- Login/logout
- Role-specific dashboards
- View/create/update/assign tickets
- Comment inside tickets
- Admins can see system logs

Simple HTML templates and CSS. Navigation bar and login redirects.

---

## API Endpoints

- `POST /login`
- `GET /tickets`
- `POST /tickets`
- `PUT /tickets/:id`
- `GET /tickets/:id/comments`
- `POST /tickets/:id/comments`
- `GET /logs` (admin only)

Use JWT in the Authorization header. REST style.

---

## How to Run

1. Clone the repo
2. Install Go (built with Go 1.24.1)
3. Open terminal, cd into project folder

Run CLI mode (default):

```bash
go run main.go
```

Run WebUI mode:

```bash
go run main.go web
```

Run and reseed database before starting CLI mode:

```bash
go run main.go seed
```

Run and reseed database before starting WebUI mode:

```bash
go run main.go web seed
```

- If you include `seed`, RyanForce will wipe and reload demo accounts, techs, clients, and tickets before starting.
- If you leave it out, it will just start normally without reseeding.

When running WebUI mode, visit [http://localhost:8080](http://localhost:8080)

Log files will show up under `logs/ryanforce.log`

---

## Seeding Demo Data

- When you run the program for the first time, RyanForce automatically seeds the database if no users are found.
- The seed will create:
  - One admin account
  - Several techs with different skill sets
  - Clients attached to mock companies
  - A mix of assigned and unassigned support tickets
- Admins can also manually wipe and reseed the database by using the `clear-db` command in the CLI.

### Seeded User Credentials

- Admin:
  - Email: `admin@example.com`
  - Password: `Admin123!`
- Techs:
  - Alice Anderson
    - Email: `alice.tech@example.com`
    - Password: `Tech123!`
  - Bob Brown
    - Email: `bob.tech@example.com`
    - Password: `Tech123!`
  - Charlie Clark
    - Email: `charlie.tech@example.com`
    - Password: `Tech123!`
- Clients:
  - Cindy Client
    - Email: `cindy.client@acme.com`
    - Password: `Client123!`
  - Gary Globex
    - Email: `gary.client@globex.com`
    - Password: `Client123!`

---

## How to Run Tests

```bash
go test ./handlers
```

- Uses a temporary in-memory database
- Tests login, role dashboards, viewing tickets, session handling

GitHub Actions also automatically runs these tests on push.

---

## Notes

- Passwords must be 8–32 characters with a capital letter, number, and special character
- Sessions expire in 24 hours
- Some features (real-time WebSockets, email alerts) are on the roadmap

---

## Credits

Built by Ryan Patton  
IFT 365 - Applied Programming Languages for IT  
Spring 2025

