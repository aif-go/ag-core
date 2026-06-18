## 1. Case-Insensitive Matching with strings.ToLower

- [x] RED: Test all casing variants (manual/Manual/MANUAL) return ManualPartitioner — config_test.go → TestToSarama_CaseInsensitive
    Assertion: `"manual"`, `"Manual"`, `"MANUAL"` all return `sarama.NewManualPartitioner` with nil error
    Expected failure: Only `"Manual"` exact match currently works; lowercase/all-caps fall to default branch

- [x] GREEN: Change constant value to lowercase + use strings.ToLower in switch — config.go → PartitionerTypeManual / ToSarama()
    References RED test: TestToSarama_CaseInsensitive
    Verification: go test -run TestToSarama_CaseInsensitive -count=1

- [x] REFACTOR: Add doc comment on ToSarama() noting case-insensitive matching — config.go

## 2. New Random and RoundRobin Partitioner Types

- [x] RED: Test random and roundrobin return correct constructors — config_test.go → TestToSarama_NewPartitioners
    Assertion: `"random"`→`sarama.NewRandomPartitioner`, `"roundrobin"`→`sarama.NewRoundRobinPartitioner`
    Expected failure: Case branches do not exist

- [x] GREEN: Add PartitionerTypeRandom/PartitionerTypeRoundRobin constants and case branches — config.go → ToSarama()
    References RED test: TestToSarama_NewPartitioners
    Verification: go test -run TestToSarama_NewPartitioners -count=1

## 3. Invalid Input Returns Error Instead of Silent Fallback

- [x] RED: Test invalid value returns error — config_test.go → TestToSarama_InvalidValue
    Assertion: `PartitionerType("unknown")` returns `nil, error` where error message contains "invalid partitioner type"
    Expected failure: Currently logs warning only and returns nil, nil (no error)

- [x] GREEN: Return fmt.Errorf in default branch — config.go → ToSarama()
    References RED test: TestToSarama_InvalidValue
    Verification: go test -run TestToSarama_InvalidValue -count=1

- [x] REFACTOR: Consolidate all ToSarama test cases into one table-driven test — config_test.go

## 4. Verification

- [x] go build ./contribute/agsarama/...
- [x] go test ./contribute/agsarama/... -count=1
