use wasm_bindgen::prelude::*;
use wasm_bindgen::JsCast;
use web_sys::{Element, ShadowRootInit, ShadowRootMode};
use yew::{create_portal, html, Component, Context, Html, NodeRef, Properties};
use mini_moka::unsync::Cache;
use chrono::{DateTime, Utc, Duration};
use md5::{Md5, Digest};
use crypto_common::{Output};
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

#[derive(Debug)]
#[derive(Eq, Hash, PartialEq)]
pub struct MsgKey(String, i64);

#[derive(Debug)]
#[derive(Eq, PartialEq, Clone)]
pub struct MsgValue(String);

pub enum MessageBoard {
    Append,
}

impl MsgKey {
    pub(crate) fn hash(&self) -> Output<Md5> {
        let mut hasher = Md5::new();
        hasher.update(self.0.as_bytes());
        hasher.update(&self.1.to_le_bytes());
        hasher.finalize()
    }

    pub(crate) fn to_string(&self) -> String {
        base16ct::lower::encode_string(&self.hash())
    }
}

pub struct TermView<MsgKey, MsgValue> {
    root_el: Element,
    messages: Cache<MsgKey, MsgValue>,
}

#[cfg(target_arch = "wasm32")]
impl Component for TermView<MsgKey, MsgValue> {
    type Message = MessageBoard;
    type Properties = ();

    fn create(_ctx: &Context<Self>) -> Self {
        let document = gloo::utils::document();
        let body = document.body().expect("body element to be present");
        let root_el = document.create_element("div").expect("root element created");

        let _ = root_el.set_id("root");
        let _ = body.append_child(&root_el);

        Self {
            root_el,
            messages: Cache::builder()
                .time_to_live(Duration::seconds(60))
                .time_to_idle(Duration::seconds(30))
                .build(),
        }
    }

    fn update(&mut self, _ctx: &Context<Self>, msg: Self::Message) -> bool {
        let bytes = vec![0x41, 0x42, 0x43];
        let utf8_string = String::from_utf8(bytes)
            .map_err(|non_utf8| String::from_utf8_lossy(non_utf8.as_bytes()).into_owned())
            .unwrap();

        let dt: DateTime<Utc> = Utc::now();

        match msg {
            MessageBoard::Append => {
                self.messages.insert(
                    MsgKey("asdasd".to_string(), dt.timestamp()),
                    MsgValue(utf8_string.to_string()),
                )
            },
        }
        true
    }

    fn view(&self, ctx: &Context<Self>) -> Html {
        let onclick = ctx.link().callback(|_| MessageBoard::Append);
        let content = create_portal(
            html! {
                {
                    for self.messages.iter().map(|(k, v)|
                        html! {
                            <div class="card w-50 card_style">
                                <div class="card-body">
                                    <p class="card-text">{format!("{:#?} / {:#?}", k.to_string(), v)}</p>
                                </div>
                            </div>
                        }
                    )
                }
            },
            self.root_el.clone(),
        );
        html! {
            <div>
                <pre>{content}</pre>
                <ShadowDOMHost>
                    <button {onclick}>{"Click me!"}</button>
                </ShadowDOMHost>
            </div>
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
    let _ = yew::Renderer::<TermView<MsgKey, MsgValue>>::new().render();

    match start_socket() {
        Ok(sock) => Ok(sock),
        Err(e) => return Err(e),
    }
}

// let window = web_sys::window().expect("no global `window` exists");
// let document = window.document().expect("should have a document on window");
// let body = document.body().expect("document should have a body");

// let iframe_el = document.create_element("iframe")?;
// body.append_child(&iframe_el)?;
