schema "billing" {}

table "loan" {
  schema = schema.billing
  column "id" {
    null = false
    type = uuid
  }
  column "currency" { # ISO 4217
    null = false
    type = char(3)
  }
  column "principal" {
    null = false
    type = integer
  }
  column "annual_interest_rate" {
    null = false
    type = integer
  }
  column "total_interest" {
    null = false
    type = integer
  }
  column "outstanding_balance" {
    null = false
    type = integer
  }
  column "start_date" {
    null = false
    type = timestamptz
  }
  column "loan_term_week" {
    null = false
    type = integer
  }
  column "weekly_payment" {
    null = false
    type = integer
  }
  column "weekly_interest" {
    null = false
    type = integer
  }
  column "created_at" {
    null    = false
    type    = timestamptz
    default = sql("now()")
  }
  column "is_completed" {
    null    = false
    type    = boolean
    default = false
  }
  primary_key {
    columns = [column.id]
  }
}

table "delinquency_status" {
  schema = schema.billing
  column "id" {
    null = false
    type = uuid
  }
  column "loan_id" {
    null = false
    type = uuid
  }
  column "is_delinquent" {
    null    = false
    type    = boolean
    default = false
  }
  column "late_fee" {
    null = true
    type = integer
  }
  column "last_payment_date" {
    null = true
    type = timestamptz
  }
  column "next_expected_payment_date" {
    null = false
    type = timestamptz
  }
  index "loan_id" {
    unique  = false
    columns = [column.loan_id]
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "loan_id_fk_delinquency" {
    columns     = [column.loan_id]
    ref_columns = [table.loan.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "payment" {
  schema = schema.billing
  column "id" {
    null = false
    type = uuid
  }
  column "loan_id" {
    null = false
    type = uuid
  }
  column "date" {
    null = false
    type = timestamptz
  }
  column "amount" {
    null = false
    type = integer
  }
  column "balance_before" {
    null = false
    type = integer
  }
  column "balance_after" {
    null = false
    type = integer
  }
  index "loan_id" {
    unique  = false
    columns = [column.loan_id]
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "loan_id_fk_payment" {
    columns     = [column.loan_id]
    ref_columns = [table.loan.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}

table "billing" {
  schema = schema.billing
  column "id" {
    null = false
    type = uuid
  }
  column "loan_id" {
    null = false
    type = uuid
  }
  column "term_number" {
    null = false
    type = integer
  }
  column "payment_due_date" {
    null = false
    type = timestamptz
  }
  column "repayment" {
    null = false
    type = integer
  }
  column "is_paid" {
    null = false
    type = boolean
    default = false
  }
  index "loan_id" {
    unique  = false
    columns = [column.loan_id]
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "loan_id_fk_payment" {
    columns     = [column.loan_id]
    ref_columns = [table.loan.column.id]
    on_update   = NO_ACTION
    on_delete   = CASCADE
  }
}