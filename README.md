# Harajuku Backend

## Description

A simple RESTful backend application for Harajuku Beauty Shop.

It uses [Gin](https://gin-gonic.com/) as the HTTP framework and [PostgreSQL](https://www.postgresql.org/) as the database with [pgx](https://github.com/jackc/pgx/) as the driver and [Squirrel](https://github.com/Masterminds/squirrel/) as the query builder. It also utilizes [Redis](https://redis.io/) as the caching layer with [go-redis](https://github.com/redis/go-redis/) as the client.

This project uses Hexagonal architecture in order to get an application that follows clean code principles.

## Getting Started

1. Ensure you have [Go](https://go.dev/dl/) 1.24 or higher and [Task](https://taskfile.dev/installation/) installed on your machine:

    ```bash
    go version && task --version
    ```

2. Create a copy of the `.env.example` file and rename it to `.env`:

    ```bash
    cp .env.example .env
    ```

    Update configuration values as needed.

3. Install all dependencies, run docker compose, create database schema, and run database migrations:

    ```bash
    task
    task service:up
    ```

4. Run the project in development mode:

    ```bash
    task dev
    ```

## Branch naming guide

Follow the [branch naming guide](./docs/branch-naming-guide.md) if you want to contribute.
