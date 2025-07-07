use wasm_bindgen::prelude::*;
use wasm_bindgen::JsCast;
use web_sys::{Element, ShadowRootInit, ShadowRootMode};
use yew::{create_portal, html, Component, Context, Html, NodeRef, Properties};
use mini_moka::sync::Cache;
// use chrono::Duration;
use crate::wsclient::start_socket;

#[derive(Properties, PartialEq)]
pub struct ShadowDOMProps {
    #[prop_or_default]
    pub children: Html,
}

pub struct ShadowDOMHost {
    host_ref: NodeRef,
    inner_host: Option<Element>,
}

impl Component for ShadowDOMHost {
    type Message = ();
    type Properties = ShadowDOMProps;

    fn create(_: &Context<Self>) -> Self {
        Self {
            host_ref: NodeRef::default(),
            inner_host: None,
        }
    }

    fn rendered(&mut self, ctx: &Context<Self>, first_render: bool) {
        if first_render {
            let shadow_root = self
                .host_ref
                .get()
                .expect("rendered host")
                .unchecked_into::<Element>()
                .attach_shadow(&ShadowRootInit::new(ShadowRootMode::Open))
                .expect("installing shadow root succeeds");
            let inner_host = gloo::utils::document()
                .create_element("div")
                .expect("can create inner wrapper");
            shadow_root
                .append_child(&inner_host)
                .expect("can attach inner host");
            self.inner_host = Some(inner_host);
            ctx.link().send_message(());
        }
    }

    fn update(&mut self, _: &Context<Self>, _: Self::Message) -> bool {
        true
    }

    fn view(&self, ctx: &Context<Self>) -> Html {
        let contents = if let Some(ref inner_host) = self.inner_host {
            create_portal(ctx.props().children.clone(), inner_host.clone())
        } else {
            html! { <></> }
        };
        html! {
            <div ref={self.host_ref.clone()}>
                {contents}
            </div>
        }
    }
}

pub type MyKey = String;
pub type MyValue = u32;

pub struct TermView<MyKey, MyValue> {
    root_el: Element,
    messages: Vec<String>,
    cache: Cache<MyKey, MyValue>,
}

pub enum TermMessage {
    Append,
}

#[cfg(target_arch = "wasm32")]
impl Component for TermView<MyKey, MyValue> {
    type Message = TermMessage;
    type Properties = ();

    fn create(_ctx: &Context<Self>) -> Self {
        let document = gloo::utils::document();
        let body = document.body().expect("body element to be present");
        let root_el = document.create_element("div").expect("root element created");

        let _ = root_el.set_id("root");
        let _ = body.append_child(&root_el);

        let cache = Cache::builder()
            // Time to live (TTL): 30 minutes
            // .time_to_live(Duration::milliseconds(2 * 60).to_std().expect("REASON"))
            // Time to idle (TTI):  5 minutes
            // .time_to_idle(Duration::milliseconds(1 * 60).to_std().expect("REASON"))
            // Create the cache.
            .build();

        Self {
            root_el,
            cache,
            messages: Vec::new(),
        }
    }

    fn update(&mut self, _ctx: &Context<Self>, msg: Self::Message) -> bool {
        let bytes = vec![0x41, 0x42, 0x43];
        let utf8_string = String::from_utf8(bytes)
            .map_err(|non_utf8| String::from_utf8_lossy(non_utf8.as_bytes()).into_owned())
            .unwrap();
        match msg {
            TermMessage::Append => {
                self.cache.insert(utf8_string.to_string(), 4);
                self.messages.push(utf8_string)
            },
        }
        true
    }

    fn view(&self, ctx: &Context<Self>) -> Html {
        let onclick = ctx.link().callback(|_| TermMessage::Append);
        let content = create_portal(
            html! {
                {format!("{}", self.messages.join("\n"))}
            },
            self.root_el.clone(),
        );
        html! {
            <>
            <div>
                <pre>{content}</pre>
                <ShadowDOMHost>
                    <button {onclick}>{"Click me!"}</button>
                </ShadowDOMHost>
            </div>
            </>
        }
    }
}

// #[wasm_bindgen]
// extern "C" {
//     #[wasm_bindgen(js_namespace = window)]
//     fn register(&c );
// }

#[wasm_bindgen(start)]
fn run() -> Result<(), JsValue> {
    // let cache = Cache::builder().expire_after(expiry).build();

    // const NUM_KEYS: usize = 10;

    // // Insert some key-value pairs.
    // for key in 0..NUM_KEYS {
    //     cache.insert(key, format!("value-{key}"));
    // }

    // // Get all entries.
    // for key in 0..NUM_KEYS {
    //     assert_eq!(cache.get(&key), Some(format!("value-{key}")));
    // }

    // // Update all entries.
    // for key in 0..NUM_KEYS {
    //     cache.insert(key, format!("new-value-{key}"));
    // }

    yew::Renderer::<TermView<MyKey, MyValue>>::new().render();
    Ok(start_socket()?)
}

// let window = web_sys::window().expect("no global `window` exists");
// let document = window.document().expect("should have a document on window");
// let body = document.body().expect("document should have a body");

// let iframe_el = document.create_element("iframe")?;
// body.append_child(&iframe_el)?;
