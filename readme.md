# Medicore HMS Backend

A robust Hospital Management System backend built with Go, Chi router, and PostgreSQL.

## Overview

Medicore HMS is a comprehensive Hospital Management System designed to streamline hospital operations, manage appointments between doctors and patients, and maintain medical records securely. This repository contains the backend API service that powers the Medicore HMS application.

## Tech Stack

- **Language**: Go
- **Web Framework**: [Chi Router](https://github.com/go-chi/chi)
- **Database**: PostgreSQL
- **API Documentation**: Swagger 2.0
- **Authentication**: JWT-based token authentication

## Features

- 🔐 **Secure Authentication System** - Register, login, and token-based authentication
- 👨‍⚕️ **Doctor Management** - Create and manage doctor profiles with specializations
- 🗓️ **Appointment Scheduling** - Book and manage appointments between doctors and patients
- 👥 **User/Patient Management** - User registration and profile management
- ⏰ **Doctor Availability** - Track and manage doctor schedules and availability
- 🔍 **User Role Management** - Different access levels for different user types

## API Endpoints

### Authentication

- `POST /v1/authentication/user` - Register a new user
- `POST /v1/authentication/token` - Create authentication token (login)

### Users

- `GET /v1/users/{id}` - Fetch a user profile by ID
- `GET /v1/users/patients` - Get all patients in the system
- `PUT /v1/users/activate/{token}` - Activate a user account via invitation token

### Doctors

- `GET /v1/doctors` - Fetch all doctors
- `POST /v1/doctors` - Create a new doctor account
- `GET /v1/doctors/{doctorID}` - Fetch a specific doctor by ID
- `POST /v1/doctors/availability` - Create doctor availability slots

### Appointments

- `GET /v1/appointments` - Get all appointments with patient and doctor information
- `POST /v1/appointments` - Create a new appointment

### System

- `GET /v1/health` - System health check endpoint

## Getting Started

### Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- Git

### Installation

1. Clone the repository

   ```bash
   git clone https://github.com/MdHasib01/hms_server
   cd medicore-hms
   ```

2. Install dependencies

   ```bash
   go mod download
   ```

3. Set up environment variables

   ```bash
   cp .env.example .env
   # Edit .env with your database credentials and other configuration
   ```

4. Set up the database

   ```bash
   # Create the database
   createdb medicore_hms

   # Run migrations (if applicable)
   go run cmd/migrate/main.go
   ```

5. Run the application
   ```bash
   go run cmd/api/main.go
   ```

The API will be available at `http://localhost:8080` (or the port specified in your configuration).

## Database Schema

The system uses the following main entities:

- **Users**: Base user accounts (patients, doctors, admins)
- **Roles**: User permission levels
- **Doctors**: Extended profile information for medical professionals
- **Appointments**: Scheduled meetings between doctors and patients
- **Availability**: Doctor's available time slots

## API Authentication

The API uses JWT token-based authentication. To access protected endpoints:

1. Register or login to get your token
2. Include the token in the Authorization header of your requests:
   ```
   Authorization: Bearer your_token_here
   ```

## Development

### Running Tests

```bash
go test ./...
```

### API Documentation

Swagger documentation is available at `/swagger/index.html` when the server is running.

You can also find the full API specification in the `swagger.json` file.

## Deployment

### Docker

```bash
# Build the Docker image
docker build -t medicore-hms .

# Run the container
docker run -p 8080:8080 --env-file .env medicore-hms
```

### Production Considerations

- Set up proper SSL/TLS certificates for HTTPS
- Configure proper database connection pooling
- Set up monitoring and logging
- Implement database backups

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the Apache 2.0 License - see the LICENSE file for details.

## Contact

For support or inquiries, please contact:

- Email: md.hasibuzzaman001@gmail.com
- Issue Tracker: [GitHub Issues](https://github.com/MdHasib01/hms_server)
