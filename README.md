# capstone-backend

**Initial setup**
- brew install --cask docker
- chmod +x run.sh
---

**Running the backend**
1. run the docker application in the backround:

2. Type this command to run program
    - ./run.sh

---

Common Docker Commands

    View logs: docker-compose logs
    Reset database: docker-compose down -v && docker-compose up -d

Troubleshooting

    If port 5433 is already in use, stop your local PostgreSQL service:

Mac/Linux: sudo service postgresql stop or brew services stop postgresql


