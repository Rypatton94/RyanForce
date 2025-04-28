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
set RYANFORCE_MODE=web   # Windows
export RYANFORCE_MODE=web # Mac/Linux
go run main.go
```

If WebUI mode: visit [http://localhost:8080](http://localhost:8080)

Log files will show up under `logs/ryanforce.log`

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

