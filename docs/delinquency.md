# Delinquency
an account is delinquent if no payment has been made for 2 terms

visualization:
```
|----------------------|----------------------|----------------------|
s                     s+1                    s+2                    s+3
```

`s` is start of the loan where the number is is weeks (e.g. `s+1` s + 1 week (7 days))

## Cases

delinquency
```
|----------------------|----------------------|----------------------|
s                     s+1                    s+2                    s+3
[                NO PAYMENT                   ][       DELINQUENT    ]
```

if no payment in `s..s+2` then after `s+2` the account will be delinquent

if payment occur in `s..s+2` they have to pay 2x for it accounted as payment (repayment)