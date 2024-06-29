schema "public" {
  comment = "standard public schema"
}
table "samples" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  column "sampleset_id" {
    null = false
    type = integer
  }
  column "case_id" {
    null = false
    type = integer
  }
  primary_key {
    columns = [column.id]
  }
  index "samples_sampleset_id_id" {
    columns = [column.sampleset_id, column.id]
  }
  foreign_key "samples_sampleset_id" {
    columns = [column.sampleset_id]
    ref_columns = [table.samplesets.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
  foreign_key "samples_case_id" {
    columns = [column.case_id]
    ref_columns = [table.cases.column.id]
    on_update = CASCADE
    on_delete = RESTRICT
  }
}
table "sample_choices" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "sample_id" {
    null = false
    type = integer
  }
  column "content" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.sample_id, column.id]
  }
  unique "sample_choices_id" {
    columns = [column.id]
  }
  foreign_key "sample_choices_sample_id" {
    columns = [column.sample_id]
    ref_columns = [table.samples.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
}
table "ratings" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "user_id" {
    null = false
    type = text
  }
  column "sample_id" {
    null = false
    type = integer
  }
  column "choice_id" {
    null = false
    type = integer
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.user_id, column.id]
  }
  unique "ratings_id" {
    columns = [column.id]
  }
  foreign_key "ratings_sample_id" {
    columns = [column.sample_id]
    ref_columns = [table.samples.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
  foreign_key "ratings_choices_id" {
    columns = [column.choice_id]
    ref_columns = [table.sample_choices.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
}
table "rating_states" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "user_id" {
    null = false
    type = text
  }
  column "rating_id" {
    null = false
    type = integer
  }
  column "state" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.user_id, column.rating_id, column.id]
  }
  foreign_key "rating_states_rating_id" {
    columns = [column.rating_id]
    ref_columns = [table.ratings.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
}
table "samplesets" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "experiment_id" {
    null = false
    type = integer
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "samplesets_experiment_id" {
    columns = [column.experiment_id]
    ref_columns = [table.experiments.column.id]
    on_update = CASCADE
    on_delete = RESTRICT
  }
}
table "experiments" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "model_id" {
    null = false
    type = text
  }
  column "prompt_id" {
    null = false
    type = text
  }
  column "user_instruction" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "experiments_model_id" {
    columns = [column.model_id]
    ref_columns = [table.models.column.id]
    on_update = CASCADE
    on_delete = RESTRICT
  }
  foreign_key "experiments_prompt_id" {
    columns = [column.prompt_id]
    ref_columns = [table.prompts.column.id]
    on_update = CASCADE
    on_delete = RESTRICT
  }
}
table "models" {
  schema = schema.public
  column "id" {
    null = false
    type = text
  }
  column "display_name" {
    null = false
    type = text
  }
  column "release_date" {
    null = false
    type = timestamptz
  }
  column "canonical_url" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  column "provider_id" {
    null = false
    type = text
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "models_provider_id" {
    columns = [column.provider_id]
    ref_columns = [table.model_providers.column.id]
    on_update = CASCADE
    on_delete = RESTRICT
  }
}
table "model_providers" {
  schema = schema.public
  column "id" {
    null = false
    type = text
  }
  column "canonical_url" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
table "prompts" {
  schema = schema.public
  column "id" {
    null = false
    type = text
  }
  column "content" {
    null = false
    type = jsonb
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
table "cases" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "caseset_id" {
    null = false
    type = integer
  }
  column "content" {
    null = false
    type = jsonb
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  column "truth" {
    null = false
    type = text
  }
  primary_key {
    columns = [column.caseset_id, column.id]
  }
  unique "cases_id" {
    columns = [column.id]
  }
  foreign_key "cases_caseset_id" {
    columns = [column.caseset_id]
    ref_columns = [table.casesets.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
}
table "case_labels" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "case_id" {
    null = false
    type = integer
  }
  column "label" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.case_id, column.id]
  }
  foreign_key "case_labels_case_id" {
    columns = [column.case_id]
    ref_columns = [table.cases.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
}
table "casesets" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.id]
  }
}
table "rating_labels" {
  schema = schema.public
  column "id" {
    null = false
    type = serial
  }
  column "rating_id" {
    null = false
    type = integer
  }
  column "label" {
    null = false
    type = text
  }
  column "create_time" {
    null = false
    type = timestamptz
  }
  primary_key {
    columns = [column.rating_id, column.id]
  }
  foreign_key "rating_labels_rating_id" {
    columns = [column.rating_id]
    ref_columns = [table.ratings.column.id]
    on_update = CASCADE
    on_delete = CASCADE
  }
}
