use std::io::{self};
use once_cell::sync::Lazy;
use fancy_regex::{Regex,Error};

static REGEX_ANSI: Lazy<Regex> = Lazy::new(|| {
    Regex::new(r"(?:\\033\[(\d+(?:;\d+)*)?([cnRhlABCDfsurgKJipm]))").unwrap()
});

pub fn to_html(input_data: &str, palette: &[&str]) -> io::Result<()> {
    let html = ansi2html(&input_data, palette).unwrap();
    println!("{:?}", html);
    Ok(())
}

fn ansi2html(text: &str, palette: &[&str]) -> Result<String, Error> {
    let mut result = String::new();

    let mut stack: Vec<String> = Vec::new();
    let mut remaining = text;
    let mut offset = 0;

    while let Ok(Some(mat)) = REGEX_ANSI.find(remaining) {
        let start = offset + mat.start();
        let end = offset + mat.end();

        result.push_str(&text[offset..start]);

        if let Ok(Some(caps)) = REGEX_ANSI.captures(&text[start..end]) {
            let codes = caps.get(1).map_or("", |m| m.as_str());
            let cmd = caps.get(2).map_or("", |m| m.as_str());

            if cmd != "m" {
                offset = end;
                remaining = &text[offset..];
                continue;
            }

            for c in codes.split(';') {
                match c {
                    "0" => {
                        while stack.pop().is_some() {
                            result.push_str("</span>");
                        }
                    }
                    c @ ("30" | "31" | "32" | "33" | "34" | "35" | "36" | "37") => {
                        let idx = c[1..].parse::<usize>().unwrap_or(0);
                        let color = palette.get(idx).unwrap_or(&"#000000");
                        result.push_str(&format!(r#"<span style="color:{}">"#, color));
                        stack.push("span".to_string());
                    }
                    _ => {}
                }
            }
        }

        offset = end;
        remaining = &text[offset..];
    }

    result.push_str(&text[offset..]);
    while stack.pop().is_some() {
        result.push_str("</span>");
    }

    Ok(result)
}
