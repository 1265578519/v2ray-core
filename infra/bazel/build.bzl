def _go_command(ctx):
  output = ctx.attr.output
  if ctx.attr.os == "windows":
    output = output + ".exe"

  output_file = ctx.actions.declare_file(ctx.attr.os + "/" + ctx.attr.arch + "/" + ctx.attr.ver + "/" + output)
  pkg = ctx.attr.pkg

  ld_flags = "-s -w -buildid="
  if ctx.attr.ld:
    ld_flags = ld_flags + " " + ctx.attr.ld

  options = [
    "go",
    "build",
    "-trimpath",
    "-o", output_file.path, 
    "-ldflags", "'%s'" % ld_flags,
    "-tags", "'%s'" % ctx.attr.gotags,
    pkg,
  ]

  command = " ".join(options)

  envs = [
    "CGO_ENABLED=0",
    "GOOS="+ctx.attr.os,
    "GOARCH="+ctx.attr.arch,
    "GO111MODULE=on",
    "GOCACHE=${TMPDIR}/gocache"
  ]

  if ctx.attr.mips: # https://github.com/golang/go/issues/27260
    envs+=["GOMIPS="+ctx.attr.mips]
    envs+=["GOMIPS64="+ctx.attr.mips]
    envs+=["GOMIPSLE="+ctx.attr.mips]
    envs+=["GOMIPS64LE="+ctx.attr.mips]
  if ctx.attr.arm:
    envs+=["GOARM="+ctx.attr.arm]

  switchToPwd="cd ${SPWD} && "

  command = switchToPwd + " ".join(envs) + " " + command

  ctx.actions.run_shell(
    outputs = [output_file],
    command = command,
    use_default_shell_env = True,
  )
  runfiles = ctx.runfiles(files = [output_file])
  return [DefaultInfo(executable = output_file, runfiles = runfiles)]


foreign_go_binary = rule(
  _go_command,
  attrs = {
    'pkg': attr.string(),
    'output': attr.string(),
    'os': attr.string(mandatory=True),
    'arch': attr.string(mandatory=True),
    'ver': attr.string(mandatory=True),
    'mips': attr.string(),
    'arm': attr.string(),
    'ld': attr.string(),
    'gotags': attr.string(),
  },
  executable = True,
)
