package java

import (
	"fmt"

	"github.com/railwayapp/railpack/core/generate"
	"github.com/railwayapp/railpack/core/plan"
)

type JavaProvider struct{}

func (p *JavaProvider) Name() string {
	return "java"
}

func (p *JavaProvider) Detect(ctx *generate.GenerateContext) (bool, error) {
	return ctx.App.HasMatch("pom.{xml,atom,clj,groovy,rb,scala,yaml,yml}") || ctx.App.HasMatch("gradlew"), nil
}

func (p *JavaProvider) Initialize(ctx *generate.GenerateContext) error {
	return nil
}

func (p *JavaProvider) StartCommandHelp() string {
	return ""
}

func (p *JavaProvider) Plan(ctx *generate.GenerateContext) error {

	build := ctx.NewCommandStep("build")
	build.AddCommand(plan.NewCopyCommand("."))
	build.Inputs = []plan.Input{plan.NewStepInput(ctx.GetMiseStepBuilder().Name())}
	var _ error

	if p.usesGradle(ctx) {
		_ = p.setGradleVersion(ctx)
		_ = p.setJDKVersion(ctx, ctx.GetMiseStepBuilder())

		if ctx.App.HasMatch("gradlew") && !ctx.App.IsFileExecutable("gradlew") {
			build.AddCommand(plan.NewExecCommand("chmod +x gradlew"))
		}

		build.AddCommand(plan.NewExecCommand("./gradlew clean build -x check -x test"))
		build.AddCache(p.gradleCache(ctx))
	} else {
		ctx.GetMiseStepBuilder().Default("maven", "latest")
		_ = p.setJDKVersion(ctx, nil)

		if ctx.App.HasMatch("mvnw") && !ctx.App.IsFileExecutable("mvnw") {
			build.AddCommand(plan.NewExecCommand("chmod +x mvnw"))
		}

		build.AddCommand(plan.NewExecCommand(fmt.Sprintf("%s -DoutputFile=target/mvn-dependency-list.log -B -DskipTests clean dependency:list install", p.getMavenExe(ctx))))
		build.AddCache(p.mavenCache(ctx))
	}

	runtimeDeps := ctx.NewMiseStepBuilder("mise:install-runtime")
	_ = p.setJDKVersion(ctx, runtimeDeps)

	outPath := "target/."
	if ctx.App.HasMatch("**/build/libs/*.jar") || p.usesGradle(ctx) {
		outPath = "."
	}

	ctx.Deploy.Inputs = []plan.Input{
		ctx.DefaultRuntimeInput(),
		plan.NewStepInput(runtimeDeps.Name(), plan.InputOptions{
			Include: runtimeDeps.GetOutputPaths(),
		}),
		plan.NewStepInput(build.Name(), plan.InputOptions{
			Include: []string{outPath},
		}),
	}
	ctx.Deploy.StartCmd = p.getStartCmd(ctx)

	return nil
}

func (p *JavaProvider) getStartCmd(ctx *generate.GenerateContext) string {
	if p.usesGradle(ctx) {
		buildGradle := p.readBuildGradle(ctx)
		return fmt.Sprintf("java $JAVA_OPTS -jar %s $(ls -1 */build/libs/*jar | grep -v plain)", getGradlePortConfig(buildGradle))
	} else if ctx.App.HasMatch("pom.xml") {
		return fmt.Sprintf("java %s $JAVA_OPTS -jar target/*jar", getMavenPortConfig(ctx))
	} else {
		return "java $JAVA_OPTS -jar target/*jar"
	}
}
