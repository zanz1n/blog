use std::{
    env::var,
    io::{Error, ErrorKind},
    str::FromStr,
};

/// Enum to identify if the app is running in a development or
/// production environment
pub enum ProcessEnv {
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
pub fn env_param<T: FromStr>(key: &str, default: Option<T>) -> T {
    let required = match default {
        None => true,
        Some(_) => false,
    };

    match var(key) {
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
