# frozen_string_literal: true

require 'gitlab-dangerfiles'

Gitlab::Dangerfiles.for_project(self) do |dangerfiles|
  dangerfiles.import_plugins
  dangerfiles.import_dangerfiles(except: %w[simple_roulette])
end
