use std::io::{Error, ErrorKind};

use sea_orm::{ConnectOptions, Database, DatabaseConnection};

pub async fn connect_to_postgres<C>(database_uri: C) -> Result<DatabaseConnection, Error>
where
    C: Into<ConnectOptions>,
{
    match Database::connect(database_uri).await {
        Ok(db) => Ok(db),
        Err(e) => Err(Error::new(ErrorKind::ConnectionRefused, e)),
    }
}
