## ADDED Requirements

### Requirement: PartitionerType enum SHALL be defined with full-lowercase values
The system SHALL define `PartitionerType` enum with all constant values in lowercase to ensure naming consistency.

#### Scenario: Enum values are all lowercase
- **WHEN** inspecting the `PartitionerType` constants
- **THEN** `PartitionerTypeHash` SHALL equal `"hash"`, `PartitionerTypeManual` SHALL equal `"manual"`, `PartitionerTypeRandom` SHALL equal `"random"`, `PartitionerTypeRoundRobin` SHALL equal `"roundrobin"`

### Requirement: ToSarama() SHALL match case-insensitively
The `ToSarama()` method SHALL normalize input via `strings.ToLower` before matching, ensuring backward compatibility with existing uppercase configurations.

#### Scenario: Lowercase input matches correctly
- **WHEN** calling `ToSarama()` with `PartitionerType("manual")`
- **THEN** it SHALL return `sarama.NewManualPartitioner` and nil error

#### Scenario: Uppercase input is still compatible
- **WHEN** calling `ToSarama()` with `PartitionerType("Manual")`
- **THEN** it SHALL return `sarama.NewManualPartitioner` and nil error

#### Scenario: All-caps input is still compatible
- **WHEN** calling `ToSarama()` with `PartitionerType("MANUAL")`
- **THEN** it SHALL return `sarama.NewManualPartitioner` and nil error

### Requirement: ToSarama() SHALL support Random and RoundRobin partitioners
The `ToSarama()` method SHALL recognize `"random"` and `"roundrobin"` as valid input values, in addition to `"hash"` and `"manual"`.

#### Scenario: Random partitioner
- **WHEN** calling `ToSarama()` with `PartitionerType("random")`
- **THEN** it SHALL return `sarama.NewRandomPartitioner` and nil error

#### Scenario: RoundRobin partitioner
- **WHEN** calling `ToSarama()` with `PartitionerType("roundrobin")`
- **THEN** it SHALL return `sarama.NewRoundRobinPartitioner` and nil error

### Requirement: Invalid partitioner type SHALL return error
When an unrecognized partitioner type is provided, `ToSarama()` SHALL return an error instead of silently falling back to the default partitioner.

#### Scenario: Invalid type returns error
- **WHEN** calling `ToSarama()` with `PartitionerType("unknown")`
- **THEN** it SHALL return `nil` and a non-nil error containing the message "invalid partitioner type"
