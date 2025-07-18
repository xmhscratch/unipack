use once_cell::sync::Lazy;
use wasm_bindgen::prelude::*;
use web_sys::{ErrorEvent, MessageEvent, WebSocket};
use chrono::{Utc};

use flatbuffers::FlatBufferBuilder;
use fbgen::schema::websock::{Message, MessageArgs, finish_message_buffer, root_as_message};

macro_rules! console_log {
    ($($t:tt)*) => (log(&format_args!($($t)*).to_string()))
}

static WS_HOST: Lazy<String> = Lazy::new(|| {
    "ws://192.168.56.55:3113/ws/default".to_string()
});

#[wasm_bindgen]
extern "C" {
    #[wasm_bindgen(js_namespace = console)]
    fn log(s: &str);
}

pub fn start_socket() -> Result<(), JsValue> {
    // Connect to an echo server
    let ws = WebSocket::new(&WS_HOST)?;
    // For small binary messages, like CBOR, Arraybuffer is more efficient than Blob handling
    ws.set_binary_type(web_sys::BinaryType::Arraybuffer);
    // create callback
    // let cloned_ws = ws.clone();
    let onmessage_callback = Closure::<dyn FnMut(_)>::new(move |e: MessageEvent| {
        // Handle difference Text/Binary,...
        if let Ok(abuf) = e.data().dyn_into::<js_sys::ArrayBuffer>() {
            console_log!("message event, received arraybuffer: {:?}", abuf);
            let array = js_sys::Uint8Array::new(&abuf);
            // let len = array.byte_length() as usize;
            // console_log!("Arraybuffer received {}bytes: {:?}", len, array.to_vec());

            let bytes: Vec<u8> = array.to_vec();
            read_message(&bytes[..]);
        }
    });
    // set message event handler on WebSocket
    ws.set_onmessage(Some(onmessage_callback.as_ref().unchecked_ref()));
    // forget the callback to keep it alive
    onmessage_callback.forget();

    let onerror_callback = Closure::<dyn FnMut(_)>::new(move |e: ErrorEvent| {
        console_log!("error event: {:?}", e);
    });
    ws.set_onerror(Some(onerror_callback.as_ref().unchecked_ref()));
    onerror_callback.forget();

    let cloned_ws = ws.clone();
    let onopen_callback = Closure::<dyn FnMut()>::new(move || {
        let mut bldr = FlatBufferBuilder::new();
        let mut bytes: Vec<u8> = Vec::new();

        // send ping
        let args = MessageArgs{
            event: 9,
            namespace: Some(bldr.create_string("default")),
            timestamp: Utc::now().timestamp(),
            message: None,
        };
        let end_offset = Message::create(&mut bldr, &args);
        finish_message_buffer(&mut bldr, end_offset);
        let finished_data = bldr.finished_data();
        bytes.extend_from_slice(finished_data);

        match cloned_ws.send_with_u8_array(&bytes[..]) {
            Ok(_) => console_log!("binary message successfully sent: {:?}", &bytes[..]),
            Err(err) => console_log!("error sending message: {:?}", err),
        }
    });
    ws.set_onopen(Some(onopen_callback.as_ref().unchecked_ref()));
    onopen_callback.forget();

    Ok(())
}
// -> (&str, u64)
fn read_message(buf: &[u8]) {
    let u = root_as_message(buf);
    println!("{:?}", u)
    // let name = u.name().unwrap();
    // let id = u.id();
    // (name, id)
}
