# Loan Billing Service
This service sole purpose is to record (bookkeeping) loan billing.

It has these functionalities:
1. Create Loan
1. Record a payment
1. Tell when is the next billing date, with the outstanding
1. Get delinquency status for a loan

## Database Design

### Loan
Data storage to record loan

## Endpoints

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