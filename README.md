- Init database

create database moneywall;
CREATE USER 'moneywall'@'localhost' IDENTIFIED BY 'moneywall';
GRANT ALL PRIVILEGES ON moneywall.* TO 'moneywall'@'localhost' IDENTIFIED BY 'moneywall';

- Run app
go run main.go