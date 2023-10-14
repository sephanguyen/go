if [ "$PROJECT_NAME" == "access-control" ]; then
    (cd stag-manabie/project-grant; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=stag-manabie-project-grant.tfplan && \
        terragrunt apply -no-color stag-manabie-project-grant.tfplan) &

    (cd stag-manabie/postgresql-roles-common; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan; \
    cd stag-manabie/postgresql-roles-lms; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan)

    (cd stag-manabie/postgresql-grant-role-common; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan; \
     cd ../postgresql-grant-role-lms; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan) &

    (cd uat-jprep/postgresql-roles; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=uat-jprep-postgresql-roles.tfplan && \
        terragrunt apply -no-color uat-jprep-postgresql-roles.tfplan && \
     cd ../postgresql-grant-role; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=uat-jprep-postgresql-grant-role.tfplan && \
        terragrunt apply -no-color uat-jprep-postgresql-grant-role.tfplan) &
    wait

    (cd prod-manabie/project-grant; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-manabie-project-grant.tfplan && \
        terragrunt apply -no-color prod-manabie-project-grant.tfplan) &
    (cd jp-partners/project-grant; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=jp-partners-project-grant.tfplan && \
        terragrunt apply -no-color jp-partners-project-grant.tfplan) &

    (cd jp-partners/postgresql-roles; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=jp-partners-postgresql-roles.tfplan && \
        terragrunt apply -no-color jp-partners-postgresql-roles.tfplan && \
     cd ../postgresql-grant-role; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=jp-partners-postgresql-grant-role.tfplan && \
        terragrunt apply -no-color jp-partners-postgresql-grant-role.tfplan) &

    (cd prod-renseikai/project-grant; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-renseikai-project-grant.tfplan && \
        terragrunt apply -no-color prod-renseikai-project-grant.tfplan) &
    wait

    (cd prod-renseikai/postgresql-roles; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-renseikai-postgresql-roles.tfplan && \
        terragrunt apply -no-color prod-renseikai-postgresql-roles.tfplan && \
     cd ../postgresql-grant-role; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-renseikai-postgresql-grant-role.tfplan && \
        terragrunt apply -no-color prod-renseikai-postgresql-grant-role.tfplan) &

    (cd prod-synersia/project-grant; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-synersia-project-grant.tfplan && \
        terragrunt apply -no-color prod-synersia-project-grant.tfplan) &

    # (cd prod-synersia/postgresql-roles; \
    #     TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-synersia-postgresql-roles.tfplan && \
    #     terragrunt apply -no-color prod-synersia-postgresql-roles.tfplan && \
    #  cd ../postgresql-grant-role; \
    #     TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-synersia-postgresql-grant-role.tfplan && \
    #     terragrunt apply -no-color prod-synersia-postgresql-grant-role.tfplan) &
    wait

    (cd prod-jprep2/postgresql-roles; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-jprep-postgresql-roles.tfplan && \
        terragrunt apply -no-color prod-jprep-postgresql-roles.tfplan && \
     cd ../postgresql-grant-role; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-jprep-postgresql-grant-role.tfplan && \
        terragrunt apply -no-color prod-jprep-postgresql-grant-role.tfplan) &
    wait

    (cd prod-tokyo/postgresql-roles-common; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan; \
    cd ../postgresql-roles-lms; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan) &
    wait

    (cd prod-tokyo/postgresql-roles-auth; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan; \
    cd ../postgresql-grant-role-auth; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan) &
    wait

    (cd prod-tokyo/postgresql-grant-role-common; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan; \
     cd ../postgresql-grant-role-lms; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
        terragrunt apply -no-color p.tfplan) &
    (cd prod-tokyo/project-grant; \
        TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=prod-tokyo-project-grant.tfplan && \
        terragrunt apply -no-color prod-tokyo-project-grant.tfplan) &
    wait
elif [[ "$PROJECT_NAME" == *"apps" ]]; then # only runs for apps module
    p="${PROJECT_NAME%-apps}" # remove -apps suffix to get the project folder name
    if [[ -d "${p}/postgresql-grant" ]]; then
      cd ${p}/postgresql-grant; \
          TERRAGRUNT_AUTO_INIT=true terragrunt plan -no-color -out=p.tfplan && \
          terragrunt apply -no-color p.tfplan
    fi
fi
