table "loan" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "principal" {
    null = false
    type = int
  }
}

table "delinquency" {
  column "id" {
    null = false
    type = uuid
  }
}

table "payment" {
  column "id" {
    null = false
    type = uuid
  }
}