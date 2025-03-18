CREATE TABLE "users" (
    "id" UUID PRIMARY KEY,
    "name" varchar NOT NULL,
    "lastName" varchar NOT NULL,
    "secondLastName" varchar NOT NULL,
    "email" varchar NOT NULL,
    "password" varchar NOT NULL
);

CREATE UNIQUE INDEX "email" ON "users" ("email");
