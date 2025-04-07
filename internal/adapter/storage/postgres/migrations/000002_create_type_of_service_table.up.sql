CREATE TABLE "TypeOfService" (
	"id" UUID NOT NULL UNIQUE,
	"name" TEXT NOT NULL,
	"price" REAL NOT NULL,
	PRIMARY KEY("id")
);
