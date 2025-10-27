schema "public" {
  comment = "Schema for dhcp2p"
}

table "nonces" {
    schema = schema.public
    column "id" {
        type = uuid
        default = sql("gen_random_uuid()")
        null = false
    }
    column "peer_id" {
        type = varchar(128)
        null = false
    }
    column "issued_at" {
        type = timestamptz
        null = false
    }
    column "expires_at" {
        type = timestamptz
        null = false
    }
    column "used" {
        type = boolean
        null = false
        default = false
    }
    column "used_at" {
        type = timestamptz
        null = true
    }

    primary_key {
        columns = [column.id]
    }
}

table "leases" {
  schema = schema.public
  column "token_id" {
    type = bigint
    null = false
  }
  column "peer_id" {
    type = varchar(128)
    null = false
  }
  column "expires_at" {
    type = timestamptz
    null = false
  }
  column "created_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }
  column "updated_at" {
    type = timestamptz
    null = false
    default = sql("now()")
  }

  primary_key {
    columns = [column.token_id]
  }

  index "idx_leases_expires_at" {
    columns = [column.expires_at]
  }
}

table "alloc_state" {
  schema = schema.public
  column "id" {
    type = serial
  }
  column "last_token_id" {
    type = bigint
    null = false
  }
  column "max_token_id" {
    type = bigint
    null = false
    default = 168162304
  }

  primary_key {
    columns = [column.id]
  }
}
