mod env;
mod error;
mod middlewares;
mod model;
mod repository;
mod routes;
mod utils;

use actix_cors::Cors;
use actix_web::{middleware, web::Data, App, HttpServer};
use env::{env_param, ProcessEnv};
use jsonwebtoken::{DecodingKey, EncodingKey};
use repository::{auth::AuthProvider, cache::CacheService, user::UserRepository};
use routes::{auth, user};
use sea_orm::{ConnectOptions, Database, DatabaseConnection};
use std::{
    fs,
    io::{Error, ErrorKind},
    time::Duration,
};
use utils::http::app_json_error_handler;

#[tokio::main]
async fn main() -> Result<(), Error> {
    let process_env = env_param::<ProcessEnv>("APP_ENV", Some(ProcessEnv::None));

    match process_env {
        ProcessEnv::None => {
            dotenvy::dotenv()
                .expect("No environment variables provided and .env file is inaccessible");

            // process_env = env_param::<ProcessEnv>("APP_ENV", None);
        }
        _ => {}
    };

    env_logger::init();

    // Retrieving app exec parameters from environment variables
    let port = env_param::<u16>("PORT", Some(8080));
    let actix_workers = env_param::<i32>("ACTIX_WORKERS", Some(-1));
    let database_uri = env_param::<String>("DATABASE_URI", None);
    let min_db_conns = env_param::<u32>("MIN_DB_CONNECTIONS", Some(1));
    let max_db_conns = env_param::<u32>("MAX_DB_CONNECTIONS", None);
    let db_conn_timeout = env_param::<u64>("DB_CONNECT_TIMEOUT", Some(5));
    let db_conn_idle_timeout = env_param::<u64>("DB_CONN_IDLE_TIMEOUT", Some(10));
    let jwt_key = env_param::<String>("JWT_KEY_FILE", None);
    let jwt_pub = env_param::<String>("JWT_PUB_FILE", None);
    let redis_uri = env_param::<String>("REDIS_URI", None);

    let jwt_key = fs::read(jwt_key)?;
    let jwt_pub = fs::read(jwt_pub)?;

    let jwt_enc_key = EncodingKey::from_ed_pem(&jwt_key)
        .or_else(|e| Err(Error::new(ErrorKind::InvalidInput, e)))?;

    let jwt_dec_key = DecodingKey::from_ed_pem(&jwt_pub)
        .or_else(|e| Err(Error::new(ErrorKind::InvalidInput, e)))?;

    let cache_service = CacheService::new(redis_uri).await?;

    let mut connection_opts = ConnectOptions::new(database_uri);

    connection_opts
        .max_connections(max_db_conns)
        .min_connections(min_db_conns)
        .connect_timeout(Duration::from_secs(db_conn_timeout))
        .idle_timeout(Duration::from_secs(db_conn_idle_timeout))
        .sqlx_logging_level(log::LevelFilter::Debug);

    let db = Database::connect(connection_opts)
        .await
        .or_else(|e| Err(Error::new(ErrorKind::ConnectionRefused, e)))?;

    // Using the box structure to allow multi-thread access to the
    // DatabaseConnection and Pool<M, W> instance.
    let db_box: &'static DatabaseConnection = Box::leak(Box::new(db));

    let cache_box: &'static CacheService = Box::leak(Box::new(cache_service));

    // Actix web config boilerplate
    let mut server = HttpServer::new(move || {
        // Setting up app middlewares
        let actix_logger = middleware::Logger::new("%{r}a %r %s %Dms").log_target("http_log");
        let actix_path_normalizer = middleware::NormalizePath::trim();

        let actix_cors = Cors::default()
            .allow_any_origin()
            .allow_any_header()
            .max_age(60 * 60); // 1 hour (Access-Control-Max-Age header)

        let user_repo = UserRepository::new(db_box);
        let user_repo = Data::new(user_repo);

        let auth_service =
            AuthProvider::new(db_box, cache_box, jwt_enc_key.clone(), jwt_dec_key.clone());
        let auth_service = Data::new(auth_service);

        let cache_service = Data::new(cache_box);

        App::new()
            .app_data(app_json_error_handler())
            .app_data(user_repo)
            .app_data(auth_service)
            .app_data(cache_service)
            .wrap(actix_logger)
            .wrap(actix_path_normalizer)
            .wrap(actix_cors)
            .service(user::get_self)
            .service(user::get_by_id)
            .service(user::create)
            .service(user::update_user)
            .service(user::delete_user)
            .service(auth::signin)
            .service(auth::signup)
            .service(auth::get_self)
    });

    if 0 < actix_workers {
        server = server.workers(actix_workers as usize);
    }

    server.bind(("0.0.0.0", port))?.run().await?;

    Ok(())
}
