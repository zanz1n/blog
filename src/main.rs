mod model;
mod utils;

use std::{
    env,
    io::{Error, ErrorKind},
    str::FromStr,
    time::Duration,
};

use actix_cors::Cors;
use actix_web::{middleware, web::Data, App, HttpServer};
use sea_orm::{ConnectOptions, Database, DatabaseConnection};

/// Enum to identify if the app is running in a development or
/// production environment
enum ProcessEnv {
    Development,
    Production,
    // When no environment is set
    None,
}

// Implementing this trait to easly obtain this enum from a string
impl FromStr for ProcessEnv {
    type Err = Error;

    fn from_str(s: &str) -> Result<Self, Self::Err> {
        if s == "DEV" || s == "DEVELOPMENT" {
            Ok(Self::Development)
        } else if s == "PROD" || s == "PRODUCTION" {
            Ok(Self::Production)
        } else {
            Err(Error::new(
                ErrorKind::InvalidData,
                "The value must be: DEVELOPMENT | DEV | PROD | PRODUCTION",
            ))
        }
    }
}

/// Function to eliminate boilerplate when retrieving and converting app
/// parameters from environment variables.
fn env_param<T: FromStr>(key: &str, default: Option<T>) -> T {
    let required = match default {
        None => true,
        Some(_) => false,
    };

    match env::var(key) {
        Ok(value) => match value.parse::<T>() {
            Ok(v) => v,
            Err(_) => {
                let err_str = format!(
                    "Environment variable {} must be valid but could not be parsed",
                    key
                );

                if !required {
                    log::error!("{err_str}");
                    default.unwrap()
                } else {
                    panic!("{err_str}")
                }
            }
        },
        Err(_) => {
            if !default.is_none() {
                default.unwrap()
            } else {
                panic!("Environment variable {} must be provided", key)
            }
        }
    }
}

async fn connect_to_postgres<C>(database_uri: C) -> Result<DatabaseConnection, Error>
where
    C: Into<ConnectOptions>,
{
    match Database::connect(database_uri).await {
        Ok(db) => Ok(db),
        Err(e) => Err(Error::new(ErrorKind::ConnectionRefused, e)),
    }
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    let mut process_env = env_param::<ProcessEnv>("APP_ENV", Some(ProcessEnv::None));

    match process_env {
        ProcessEnv::None => {
            dotenvy::dotenv()
                .expect("Environment variables provided and .env file is inaccessible");

            process_env = env_param::<ProcessEnv>("APP_ENV", None);
        }
        _ => {}
    };

    env_logger::init();

    // Retrieving app exec parameters from environment variables
    let port = env_param::<u16>("PORT", Some(8080));
    let actix_workers = env_param::<usize>("ACTIX_WORKERS", Some(4));
    let database_uri = env_param::<String>("DATABASE_URI", None);
    let min_db_conns = env_param::<u32>("MIN_DB_CONNECTIONS", Some(1));
    let max_db_conns = env_param::<u32>("MAX_DB_CONNECTIONS", None);
    let db_conn_timeout = env_param::<u64>("DB_CONNECT_TIMEOUT", Some(5));
    let db_conn_idle_timeout = env_param::<u64>("DB_CONN_IDLE_TIMEOUT", Some(10));

    let mut connection_opts = ConnectOptions::new(database_uri).to_owned();

    connection_opts
        .max_connections(max_db_conns)
        .min_connections(min_db_conns)
        .connect_timeout(Duration::from_secs(db_conn_timeout))
        .idle_timeout(Duration::from_secs(db_conn_idle_timeout));

    match process_env {
        ProcessEnv::Development => {
            connection_opts.sqlx_logging_level(log::LevelFilter::Debug);
        }
        _ => {
            connection_opts.sqlx_logging_level(log::LevelFilter::Info);
        }
    };

    let db = connect_to_postgres(connection_opts).await?;

    // Using the box structure to allow multi-thread access to the
    // DatabaseConnection instance.
    let db_box: &DatabaseConnection = Box::leak(Box::new(db));

    // Actix web config boilerplate
    HttpServer::new(move || {
        // Setting up app middlewares
        let actix_logger = middleware::Logger::new("%{r}a %r %s %D").log_target("http_log");
        let actix_path_normalizer = middleware::NormalizePath::trim();

        let actix_cors = Cors::default()
            .allow_any_origin()
            .allow_any_header()
            .max_age(60 * 60); // 1 hour (Access-Control-Max-Age header)

        let db_data = Data::new(db_box);

        App::new()
            .app_data(db_data)
            .wrap(actix_logger)
            .wrap(actix_path_normalizer)
            .wrap(actix_cors)
    })
    .workers(actix_workers)
    .bind(format!("0.0.0.0:{}", port))?
    .run()
    .await?;

    Ok(())
}
