@critical
Feature: Network policy enforcement

    Background:
        Given network policy is enabled in current kubernetes cluster

    Scenario Outline: Pods can access allowlisted pods
        Then pod in "<origin namespace>" namespace "can" access pod in "<target namespace>" namespace

        Examples:
            | origin namespace | target namespace |
            | backend          | backend          |
            | backend          | elastic          |
            | backend          | kafka            |
            | backend          | nats-jetstream   |
            | backend          | unleash          |
            | istio-system     | backend          |
            | elastic          | backend          |
            | monitoring       | backend          |

    # Scenario Outline: Pod cannot access non-allowlisted pods
    #     Then pod in "<origin namespace>" namespace "cannot" access pod in "<target namespace>" namespace

    #     Examples:
    #         | origin namespace | target namespace |
    #         | kube-system      | backend          |
    #         | kafka            | backend          |
