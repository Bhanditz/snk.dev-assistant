function(test)

  project_open(".")
  ans(project)

  assign(project.package_descriptor.id = 'somepkg')
  timer_start(project_write)
  project_write("${project}")
  ans(project_file_path)
  timer_print_elapsed(project_write)
  assert(EXISTS "${project_file_path}")
  assert("${project_file_path}" STREQUAL "${test_dir}/.cps/project.scmake")
  fread_data("${project_file_path}")
  ans(data)
  assertf({data.package_descriptor.id} STREQUAL "somepkg")
  project_state_assert("${project}" closed)
endfunction()