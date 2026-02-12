Feature: roam-cli q command
  As a user
  I want to query Roam API through CLI
  So that I can get results in JSON

  Scenario: Execute q against mock server
    Given a mock Roam API server
    And roam-cli env is configured for the mock server
    When I run command "q [:find ?title :where [?e :node/title ?title]]"
    Then the command should succeed
    And stdout should contain "Hello BDD"
