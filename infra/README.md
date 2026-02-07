# Build prost db
docker-compose up

# To build User service

docker build -f services/users/Dockerfile -t prost-users:latest .

# Rollback all
go run main.go -db "postgresql://prost_admin:prost_password@localhost:5432/prost?sslmode=disable" -path "./db/" -direction down

For GCP
go run main.go -db "postgresql://prost_admin:prost_password@35.209.92.214:5432/prost?sslmode=disable" -path "./db/" -direction down

# Run all migrations up
go run main.go -db "postgresql://prost_admin:prost_password@localhost:5432/prost?sslmode=disable" -path "./db/" -direction up

For GCP
go run main.go -db "postgresql://prost_admin:prost_password@35.209.92.214:5432/prost?sslmode=disable" -path "./db/" -direction up

# Verify
docker exec prost-postgres psql -U prost_admin -d prost -c "\dn"

# Enable Rabbitmq management plugin
rabbitmq-plugins enable rabbitmq_management

# Golden Rules
Rule	                        Why
Never edit applied migrations	Can't re-apply them, breaks consistency
Always use ALTER for changes	Preserves data, allows rollback
Never drop in production	    Data loss = disaster
Create new migration files	    Maintains audit trail of changes
Test down/up locally first	    Catch errors before production
Backup before production deploy	Safety net for emergency rollback
One migration per feature	    Clear git history, easier to debug