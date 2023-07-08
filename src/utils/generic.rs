use std::time::{SystemTime, UNIX_EPOCH};

pub fn now_unix() -> u128 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis()
}
