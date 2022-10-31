# smsk-bulrush-test

# Pre-reqs

## Create an ECR repo smsk-bulrush-test

$ aws-okta exec ops-write -- aws ecr create-repository --repository-name=smsk-bulrush-test

## Add permissions to ecr repo

$ aws-okta exec ops-write -- aws ecr set-repository-policy \
  --registry-id="528451384384" \
  --repository-name=smsk-bulrush-test \
  --policy-text='{"Version": "2008-10-17", "Statement": [{"Action": ["ecr:GetDownloadUrlForLayer", "ecr:BatchGetImage", "ecr:BatchCheckLayerAvailability", "ecr:DescribeImages"], "Principal": "*", "Effect": "Allow", "Condition": {"ForAnyValue:StringLike": {"aws:PrincipalOrgPaths": "o-wb8ewg4hzk/*"}}, "Sid": "AllowOrganization"}]}'
