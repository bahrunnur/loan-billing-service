# Achitecture

I use hexagonal architecture ([ref](https://www.wikiwand.com/en/articles/Hexagonal_architecture_(software))) to structure
the code that separate actual business/domain logic with logic that interact with the external

Basically:
- ports: interface to and from the business domain
- adapters: implementation of that ports
- domain/business logic: actual codebase logic
