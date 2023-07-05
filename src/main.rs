use std::{
    env,
    io::{Error, ErrorKind},
    str::FromStr,
};

use actix_cors::Cors;
use actix_web::{middleware, App, HttpServer};
use sea_orm::{ConnectOptions, Database, DatabaseConnection};

enum ProcessEnv {
    Development,
    Production,
    None,
}

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

    let port = env_param::<u16>("PORT", Some(8080));
    let actix_workers = env_param::<usize>("ACTIX_WORKERS", Some(4));
    let database_uri = env_param::<String>("DATABASE_URI", None);

    let mut db = connect_to_postgres(database_uri).await?;

    match process_env {
        ProcessEnv::Development => db.set_metric_callback(|i| {
            log::info!(target: "query_log", "Statement: {}\n\tTook {}ms {}",
                i.statement,
                i.elapsed.as_millis(),
                if i.failed { "Success" } else { "Fail" }
            );
        }),
        _ => {}
    }

    HttpServer::new(|| {
        let actix_logger = middleware::Logger::new("%{r}a %r %s %D").log_target("http_log");
        let actix_path_normalizer = middleware::NormalizePath::new(middleware::TrailingSlash::Trim);

        let actix_cors = Cors::default()
            .allow_any_origin()
            .allow_any_header()
            .max_age(60 * 60);

        App::new()
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
