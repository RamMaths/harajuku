CREATE TYPE "quote_state_enum" AS ENUM ('pending', 'approved', 'rejected', 'requires_proof');

CREATE TABLE "Quote" (
	"id" UUID NOT NULL UNIQUE,
	"typeOfServiceId" UUID NOT NULL,
	"clientId" UUID NOT NULL,
	"time" TIMESTAMPTZ NOT NULL,
	"description" TEXT,
	"state" quote_state_enum NOT NULL,
	"price" REAL NOT NULL,
    
	PRIMARY KEY("id"),
	FOREIGN KEY("typeOfServiceId") REFERENCES "TypeOfService"("id") ON UPDATE CASCADE ON DELETE RESTRICT,
	FOREIGN KEY("clientId") REFERENCES "users"("id") ON UPDATE CASCADE ON DELETE RESTRICT
);
