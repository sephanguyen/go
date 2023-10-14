package constants

var (
	ESConversationIndexName          = "conversations"
	ElasticsearchConversationMapping = `
{
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                    "access_path_prefix": {
                        "tokenizer": "path_hierarchy"
                    },
                    "kuromoji_normalize": {
                        "char_filter": [
                            "icu_normalizer"
                        ],
                        "tokenizer": "kuromoji_tokenizer",
                        "mode": "search",
                        "filter": [
                            "lowercase",
                            "edge_ngram"
                        ]
                    },
                    "english_normalize": {
                        "tokenizer": "standard",
                        "filter": [
                            "lowercase",
                            "edge_ngram"
                        ]
                    }
                },
                "filter": {
                    "edge_ngram": {
                        "type": "edge_ngram",
                        "min_gram": "1",
                        "max_gram": "25",
                        "token_chars": [
                            "letter",
                            "digit"
                        ]
                    },
                    "1_2_grams": {
                        "type": "ngram",
                        "min_gram": 1,
                        "max_gram": 2
                    }
                }
            }
        }
    },
    "mappings": {
        "_field_names": {
            "enabled": false
        },
        "properties": {
            "conversation_id": {
                "type": "keyword"
            },
            "conversation_name": {
                "properties": {
                    "english": {
                        "type": "text",
                        "analyzer": "english_normalize",
                        "search_analyzer": "standard"
                    },
                    "japanese": {
                        "type": "text",
                        "analyzer": "kuromoji_normalize"
                    }
                }
            },
            "last_message": {
                "properties": {
                    "updated_at": {
                        "type": "date"
                    }
                }
            },
            "is_replied": {
                "type": "boolean"
            },
            "owner": {
                "type": "keyword"
            },
            "conversation_type": {
                "type": "keyword"
            },
            "resource_path": {
                "type": "keyword"
            },
            "access_paths": {
                "type": "text",
                "analyzer": "access_path_prefix"
            }
        }
    }
}
`
)
