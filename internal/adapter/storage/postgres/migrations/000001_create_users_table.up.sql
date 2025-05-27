CREATE TYPE "users_role_enum" AS ENUM ('admin', 'client');

CREATE TABLE "users" (
    "id" UUID PRIMARY KEY,
    "name" varchar NOT NULL,
    "lastName" varchar NOT NULL,
    "secondLastName" varchar,
    "email" varchar NOT NULL,
    "password" varchar NOT NULL,
    "role" users_role_enum DEFAULT 'client'
);

CREATE UNIQUE INDEX "email" ON "users" ("email");
