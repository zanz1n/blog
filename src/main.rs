mod db;
mod env;
mod model;
mod repository;
mod routes;
mod utils;

use std::{io::Error, time::Duration};

use actix_cors::Cors;
use actix_web::{middleware, web::Data, App, HttpServer};
use db::connect_to_postgres;
use env::{env_param, ProcessEnv};
use repository::user::UserRepository;
use routes::user;
use sea_orm::{ConnectOptions, DatabaseConnection};
use utils::http::app_json_error_handler;

#[tokio::main]
async fn main() -> Result<(), Error> {
    let process_env = env_param::<ProcessEnv>("APP_ENV", Some(ProcessEnv::None));

    match process_env {
        ProcessEnv::None => {
            dotenvy::dotenv()
                .expect("Environment variables provided and .env file is inaccessible");

            // process_env = env_param::<ProcessEnv>("APP_ENV", None);
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
        .idle_timeout(Duration::from_secs(db_conn_idle_timeout))
        .sqlx_logging_level(log::LevelFilter::Debug);

    let db = connect_to_postgres(connection_opts).await?;

    // Using the box structure to allow multi-thread access to the
    // DatabaseConnection instance.
    let db_box: &'static DatabaseConnection = Box::leak(Box::new(db));

    // Actix web config boilerplate
    HttpServer::new(move || {
        // Setting up app middlewares
        let actix_logger = middleware::Logger::new("%{r}a %r %s %Dms").log_target("http_log");
        let actix_path_normalizer = middleware::NormalizePath::trim();

        let actix_cors = Cors::default()
            .allow_any_origin()
            .allow_any_header()
            .max_age(60 * 60); // 1 hour (Access-Control-Max-Age header)

        let user_repo = UserRepository::new(db_box);

        let user_repo = Data::new(user_repo);

        App::new()
            .app_data(app_json_error_handler())
            .app_data(user_repo)
            .wrap(actix_logger)
            .wrap(actix_path_normalizer)
            .wrap(actix_cors)
            .service(user::get_by_id)
            .service(user::create)
    })
    .workers(actix_workers)
    .bind(format!("0.0.0.0:{}", port))?
    .run()
    .await?;

    Ok(())
}
