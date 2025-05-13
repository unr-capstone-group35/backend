# capstone-backend

**Initial setup**

- Build Docker image
```
docker build . --tag capstone-backend
```

---

**Running the backend**

- Start the backend and database containers
```
docker compose up
```
---

Common Docker Commands

    View logs: docker-compose logs
    Reset database: docker-compose down -v && docker-compose up -d

Troubleshooting

    If port 5433 is already in use, stop your local PostgreSQL service:

Mac/Linux: sudo service postgresql stop or brew services stop postgresql

**Database Diagram**

![db diagram](db_diagram.svg)
