# Loan Billing Service
This service sole purpose is to record (bookkeeping) loan billing.

It has these functionalities:
1. Create Loan
1. Record a payment
1. Tell when is the next billing date, with the outstanding
1. Get delinquency status for a loan

## Database Design
The choice of database is really depending on how this service act, if it is an analytical one then it should use
columnar storage like Apache Parquet. Other than that, an ordinary table storage like PostgreSQL 
is enough to provide consistency. And, can be tuned to Availability with eventual consistency, or Partition Tolerance
with multi-node.

### Loan
Data storage to record loan

### Payments
Data storage that record payment that has been made to a loan (referenced by: `loanID`)

relation: 1 loan ..has.. n payments [1..n]

### Delinquency Status
Data storage that act as a metadata for loan delinquency status (referenced by: `loanID`)

relation: 1 loan ..has.. 1 delinquency status [1..1]

## Endpoints
The service open up some ports through gRPC, as I assume these subroutines are not accessible to the end user. But, it
act as a microservice that sole purpose is to bookkeep the loan billing.

### 1. Create Loan
Peeking at "Example 3", I assume the loan is already in `disbursed` status so a call to this only for bookkeeping

```
POST /billing/loans
```

### 1. Record a Payment
I assume the payment is being settled in other place, so a call to this functionality only for bookkeeping

```
POST /billing/loans/:id/payments
```

### 2. Billing
Return billing date and the amount with outstanding payment

```
GET /billing/loans/:id/billing
```

### 3. Delinquency Status
Return the delinquency status for a loan id

```
GET /billing/loans/:id/delinquency
```