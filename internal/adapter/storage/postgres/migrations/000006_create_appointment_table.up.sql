CREATE TYPE "appointment_status_enum" AS ENUM ('booked', 'cancelled', 'completed');

CREATE TABLE "Appointment" (
	"id" UUID NOT NULL UNIQUE,
	"clientId" UUID NOT NULL,
	"slotId" UUID NOT NULL,
	"quoteId" UUID NOT NULL UNIQUE,
	"status" appointment_status_enum NOT NULL DEFAULT 'booked',
	PRIMARY KEY("id"),
	FOREIGN KEY("clientId") REFERENCES "users"("id") ON UPDATE CASCADE ON DELETE RESTRICT,
	FOREIGN KEY("slotId") REFERENCES "AvailabilitySlot"("id") ON UPDATE CASCADE ON DELETE RESTRICT,
	FOREIGN KEY("quoteId") REFERENCES "Quote"("id") ON UPDATE CASCADE ON DELETE RESTRICT
);
