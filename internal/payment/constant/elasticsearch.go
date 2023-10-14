package constant

const (
	ElasticOrderTableName     = "order"
	ElasticOrderItemTableName = "order_item"
	ElasticProductTableName   = "product"

	ElasticOrderIndexMapping = `
{
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                },
                "filter": {
                }
            }
        }
    },
    "mappings": {
        "_field_names": {
            "enabled": false
        },
        "properties": {
            "order_id": {
                "type": "keyword"
            },
            "student_id": {
                "type": "keyword"
            },
            "student_full_name": {
                "type": "keyword"
            },
            "location_id": {
                "type": "keyword"
            },
            "order_sequence_number": {
                "type": "integer"
            },
            "order_comment": {
                "type": "text"
            },
            "order_status": {
                "type": "keyword"
            },
            "order_type": {
                "type": "keyword"
            },
            "updated_at": {
                "type": "date"
            },
            "created_at": {
                "type": "date"
            },
            "resource_path": {
                "type": "keyword"
            },
            "is_reviewed": {
                "type": "boolean"
            }
        }
    }
}
`
	ElasticOrderItemIndexMapping = `
{
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                },
                "filter": {
                }
            }
        }
    },
    "mappings": {
        "_field_names": {
            "enabled": false
        },
        "properties": {
            "order_id": {
                "type": "keyword"
            },
            "product_id": {
                "type": "keyword"
            },
            "order_item_id": {
                "type": "keyword"
            },
            "product_name": {
                "type": "text"
            },
            "discount_id": {
                "type": "keyword"
            },
            "start_date": {
                "type": "date"
            },
            "created_at": {
                "type": "date"
            },
            "resource_path": {
                "type": "keyword"
            }
        }
    }
}
`
	ElasticProductIndexMapping = `
{
    "settings": {
        "index": {
            "analysis": {
                "analyzer": {
                },
                "filter": {
                }
            }
        }
    },
    "mappings": {
        "_field_names": {
            "enabled": false
        },
        "properties": {
            "product_id": {
                "type": "keyword"
            },
            "name": {
                "type": "text"
            },
            "product_type": {
                "type": "keyword"
            },
            "tax_id": {
                "type": "keyword"
            },
            "available_from": {
                "type": "date"
            },
            "available_until": {
                "type": "date"
            },
            "custom_billing_period": {
                "type": "date"
            },
            "billing_schedule_id": {
                "type": "keyword"
            },
            "disable_pro_rating_flag": {
                "type": "boolean"
            },
            "billing_ratio_type_id": {
                "type": "keyword"
            },
            "remarks": {
                "type": "text"
            },
            "is_archived": {
                "type": "boolean"
            },
            "start_date": {
                "type": "date"
            },
            "created_at": {
                "type": "date"
            },
            "resource_path": {
                "type": "keyword"
            }
        }
    }
}
`
)
