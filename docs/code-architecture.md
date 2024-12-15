# Achitecture

I use hexagonal architecture ([ref](https://www.wikiwand.com/en/articles/Hexagonal_architecture_(software))) to structure
the code that separate actual business/domain logic with logic that interact with the external

Basically:
- ports: interface to and from the business domain
- adapters: implementation of that ports, be it to outside, or from the outside
- domain/business logic: actual codebase logic

## Folder Structure

```
├── Makefile            (makefile with targets for building binaries, proto, and test)
├── README.md           (how to)
├── cmd                 (application packages [executables])
    ├── client          (grpc client)
│   └── loanbilling     (the service binary)
├── db                  (database migration related files)
├── docs                (additional documentation)
├── internal            (internal codes that are not accessible to other service)
│   ├── adapters        (code that interacts directly with external systems)
│   ├── config          (configuration struct)
│   ├── model           (structs used internally by the service business logic)
│   ├── ports           (interface definitions for code that interacts with external systems)
│   ├── service         (configures and runs the service)
│   └── ...             (other packages containing business logic for the service)
├── pkg                 (service codes that can be used for other service)
└── proto               (protobuf stuff)
    ├── gen             (generated proto file)
    ├── loanbilling     (proto specification)
    └── buf.gen.yaml    (buf gen config: https://buf.build/docs/configuration/v2/buf-gen-yaml/)

```
