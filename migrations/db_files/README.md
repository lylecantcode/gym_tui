# Migrations 

## Warnings
Migrations are non-reversible in their current implementation.  
The idea of this, is that a change can be reverted, via another "up" migration anyway.  

## Usage
* Create a new file starting with a 3 digit number (e.g. 001 or 123).  
* Put raw sql into this file.  
* When the code is run, the migration should automatically apply.  
* Changing old migrations will not effect anything, unless the database is deleted or the `schema_migrations` table is altered.
