use scraper::{Html, Node};
use serde::{Deserialize, Serialize};
use std::io::{Error, ErrorKind};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum NodeKind {
    #[serde(rename = "h2")]
    H2,
    #[serde(rename = "h3")]
    H3,
    #[serde(rename = "h4")]
    H4,
    #[serde(rename = "h5")]
    H5,
}

impl TryFrom<&str> for NodeKind {
    type Error = Error;

    fn try_from(value: &str) -> Result<Self, Self::Error> {
        match value {
            "h2" => Ok(Self::H2),
            "h3" => Ok(Self::H3),
            "h4" => Ok(Self::H4),
            "h5" => Ok(Self::H5),
            _ => Err(Error::new(
                ErrorKind::InvalidInput,
                "Invalid heading string",
            )),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HeadingNode {
    pub title: String,
    pub kind: NodeKind,
}

pub fn get_headings(s: &str) -> Vec<HeadingNode> {
    let dom = Html::parse_fragment(s);

    let mut before_was_heading = false;
    let mut current_heading_kind = NodeKind::H2;

    let mut headings = Vec::<HeadingNode>::new();

    for ele in dom.tree {
        match ele {
            Node::Element(tag) => match NodeKind::try_from(tag.name()) {
                Ok(v) => {
                    before_was_heading = true;
                    current_heading_kind = v
                }
                Err(_) => before_was_heading = false,
            },
            Node::Text(s) => {
                if before_was_heading {
                    headings.push(HeadingNode {
                        title: s.to_string(),
                        kind: current_heading_kind.clone(),
                    })
                }
            }

            _ => before_was_heading = false,
        }
    }

    headings
}

pub fn get_first_paragraph(s: &str) -> Option<String> {
    let dom = Html::parse_fragment(s);

    let mut before_was_p = false;

    for ele in dom.tree {
        match ele {
            Node::Element(tag) => {
                if tag.name() == "p" {
                    before_was_p = true
                }
            }
            Node::Text(s) => {
                if before_was_p {
                    return Some(s.to_string());
                }
            }
            _ => before_was_p = false,
        }
    }

    None
}
