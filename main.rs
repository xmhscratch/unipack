use clap::Parser;
use lazy_static::lazy_static;
use std::collections::HashMap;
use std::fs;
use std::io::{self, Read};
use viewer::ansi::to_html;

lazy_static! {
    static ref MAP_ANSI_PALETTE: HashMap<&'static str, [&'static str; 16]> = {
        let mut m = HashMap::new();
        m.insert("console", [
            "#000000", "#AA0000", "#00AA00", "#AA5500",
            "#0000AA", "#AA00AA", "#00AAAA", "#AAAAAA",
            "#555555", "#FF5555", "#55FF55", "#FFFF55",
            "#5555FF", "#FF55FF", "#55FFFF", "#FFFFFF"
        ]);
        m
    };
}

#[derive(Parser)]
#[command(version, about, long_about = None)]
struct Args {
    /// Input file (or stdin if omitted)
    input: Option<String>,

    /// Output file (stdout if omitted)
    #[arg(short, long)]
    output: Option<String>,

    /// Palette name
    #[arg(long, default_value = "console")]
    palette: String,
}

fn main() -> io::Result<()> {
    let args = Args::parse();

    let input_data = if let Some(path) = args.input {
        fs::read_to_string(path)?
    } else {
        let mut buf = String::new();
        io::stdin().read_to_string(&mut buf)?;
        buf
    };
    let palette = MAP_ANSI_PALETTE.get(args.palette.as_str()).unwrap_or(&MAP_ANSI_PALETTE["console"]);

    let html = to_html(&input_data, palette).unwrap();

    println!("{:?}", html);
    Ok(())
}
