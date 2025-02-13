use std::{thread, time};

fn main() {
    loop {
        thread::sleep(time::Duration::from_millis(1000));
        println!("Hello");
    }
}
