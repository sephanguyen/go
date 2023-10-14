package skaffoldwrapper

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	vr "github.com/manabie-com/backend/internal/golibs/variants"

	skaffoldschema "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest"
	skaffoldutil "github.com/GoogleContainerTools/skaffold/pkg/skaffold/util"
	skaffoldv2schema "github.com/GoogleContainerTools/skaffold/v2/pkg/skaffold/schema/latest"
	skaffoldv2util "github.com/GoogleContainerTools/skaffold/v2/pkg/skaffold/util"
	"gopkg.in/yaml.v3"
	istioScheme "istio.io/client-go/pkg/apis/networking/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
	vpaScheme "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	k8sScheme "k8s.io/client-go/kubernetes/scheme"
)

type Command struct {
	es EnvSet
	fs FlagSet
}

// New returns a new, empty Command.
func New() *Command {
	return &Command{}
}

// FlagSet sets the entire flag set for this command.
func (c *Command) FlagSet(fs FlagSet) *Command {
	c.fs = fs
	return c
}

// EnvSet sets the entire environment variables set for this command.
func (c *Command) EnvSet(es EnvSet) *Command {
	c.es = es
	return c
}

// E sets `ENV` environment variable for this command.
func (c *Command) E(e vr.E) *Command {
	c.es.Env = e.String()
	return c
}

// P sets `ORG` environment variable for this command.
func (c *Command) P(p vr.P) *Command {
	c.es.Org = p.String()
	return c
}

// Filename sets the value for skaffold's --filename flag.
func (c *Command) Filename(f string) *Command {
	c.fs.F = f
	return c
}

// Profiles set the value for skaffold's --profile flag.
func (c *Command) Profile(p string) *Command {
	c.fs.P = p
	return c
}

// Profiles set the value for skaffold's --profile-auto-activation flag.
func (c *Command) ProfileAutoActivation(b bool) *Command {
	c.fs.ProfileAutoActivation = &b
	return c
}

// Diagnose runs `skaffold diagnose` but also further expand the resultant
// env template.
func (c *Command) Diagnose() (res []skaffoldschema.SkaffoldConfig, err error) {
	cmd, err := command("diagnose", c.fs.Args("--yaml-only")...)
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	cmd.Dir = execwrapper.RootDirectory()
	cmd.Env = c.es.Environ()
	cmdOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	output, err := skaffoldutil.ExpandEnvTemplate(string(cmdOutput), c.es.EnvironMap())
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(output)
	dec := yaml.NewDecoder(r)
	for {
		one := skaffoldschema.SkaffoldConfig{}
		err = dec.Decode(&one)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, one)
	}
	return res, nil
}

// V2Diagnose is similar to Diagnose but uses skaffold v2.
func (c *Command) V2Diagnose() (res []skaffoldv2schema.SkaffoldConfig, err error) {
	cmd, err := command2("diagnose", c.fs.Args("--yaml-only", "--verbosity=error")...)
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	cmd.Dir = execwrapper.RootDirectory()
	cmd.Env = c.es.Environ()
	cmdOutput, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	output, err := skaffoldv2util.ExpandEnvTemplate(string(cmdOutput), c.es.EnvironMap())
	if err != nil {
		return nil, err
	}
	r := strings.NewReader(output)
	dec := yaml.NewDecoder(r)
	for {
		one := skaffoldv2schema.SkaffoldConfig{}
		err = dec.Decode(&one)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		res = append(res, one)
	}
	return res, nil
}

// renderScriptPath is substituted in unit test for mocking purposes.
var renderScriptPath = "./scripts/render.bash"

// RenderRaw is similar to Render, but returns the raw output instead of continuing to
// parse it.
func (c *Command) RenderRaw() ([]byte, error) {
	cmd := exec.Command(renderScriptPath, c.fs.Args()...) //nolint:gosec
	cmd.Stderr = os.Stderr
	cmd.Dir = execwrapper.RootDirectory()
	cmd.Env = c.es.Environ()
	return cmd.Output()
}

// V2RenderRaw is simlar to V2, but uses skaffold v2.
func (c *Command) V2RenderRaw() ([]byte, error) {
	tmpf, err := os.CreateTemp("", fmt.Sprintf("%s_%s_%d_*.yaml", c.es.Env, c.es.Org, time.Now().Unix()))
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary output file: %s", err)
	}
	cmd, err := command2("render", c.fs.Args("--offline=true", "--digest-source=none", "--verbosity=error", "--output="+tmpf.Name())...)
	if err != nil {
		return nil, err
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Dir = execwrapper.RootDirectory()
	cmd.Env = c.es.Environ()
	if err = cmd.Run(); err != nil {
		return nil, err
	}

	return io.ReadAll(tmpf)
}

// Render calls `skaffold render` with the provided flags and env variables.
// It also further parses the output and return the proper Kubernetes objects.
//
// See CachedRender for using Render with cache.
func (c *Command) Render() ([]interface{}, error) {
	output, err := c.RenderRaw()
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %s", err)
	}

	// parse output
	b := bufio.NewReader(bytes.NewReader(output))
	r := k8sYaml.NewYAMLReader(b)
	res := []interface{}{}
	for {
		doc, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read from output: %s", err)
		}
		doc, err = k8sYaml.ToJSON(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert manifest to JSON: %s (full manifest: %q)", err, doc)
		}
		if string(doc) == "null" {
			continue
		}
		obj, err := parseK8sObjects(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubernetes object: %s, data: %q", err, doc)
		}
		res = append(res, obj)
	}
	return res, nil
}

// V2Render is similar to Render, but uses skaffold v2.
func (c *Command) V2Render() ([]interface{}, error) {
	output, err := c.V2RenderRaw()
	if err != nil {
		return nil, fmt.Errorf("failed to run command: %s", err)
	}

	// parse output
	b := bufio.NewReader(bytes.NewReader(output))
	r := k8sYaml.NewYAMLReader(b)
	res := []interface{}{}
	for {
		doc, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read from output: %s", err)
		}
		doc, err = k8sYaml.ToJSON(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert manifest to JSON: %s (full manifest: %q)", err, doc)
		}
		if string(doc) == "null" {
			continue
		}
		obj, err := parseK8sObjects(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to parse kubernetes object: %s, data: %q", err, doc)
		}
		res = append(res, obj)
	}
	return res, nil
}

// cachedRender is similar to CachedRender, but also reports
// whether cache hits.
func (c *Command) cachedRender() ([]interface{}, error, bool) { //nolint:revive
	cache, exists := globalRenderCache.Get(*c)
	if exists {
		return cache.res, cache.err, true
	}

	res, err := c.Render()
	globalRenderCache.Set(*c, renderResult{res: res, err: err})
	return res, err, false
}

// CachedRender calls Render, but also caches the result to
// a global registry for future identical calls. This is usually
// used in testings.
func (c *Command) CachedRender() ([]interface{}, error) {
	res, err, _ := c.cachedRender()
	return res, err
}

// v2CachedRender is similar to V2CachedRender, but also reports
// whether cache hits.
func (c *Command) v2CachedRender() ([]interface{}, error, bool) { //nolint:revive
	cache, exists := globalRenderCacheV2.Get(*c)
	if exists {
		return cache.res, cache.err, true
	}

	res, err := c.V2Render()
	globalRenderCacheV2.Set(*c, renderResult{res: res, err: err})
	return res, err, false
}

// V2CachedRender is similar to CachedRender, but uses skaffold v2.
func (c *Command) V2CachedRender() ([]interface{}, error) {
	res, err, _ := c.v2CachedRender()
	return res, err
}

var k8sCodec serializer.CodecFactory

func init() {
	s := runtime.NewScheme()
	if err := k8sScheme.AddToScheme(s); err != nil {
		panic(fmt.Errorf("failed to registry k8s scheme: %s", err))
	}
	if err := istioScheme.AddToScheme(s); err != nil {
		panic(fmt.Errorf("failed to registry istio scheme: %s", err))
	}
	if err := vpaScheme.AddToScheme(s); err != nil {
		panic(fmt.Errorf("failed to registry vpa scheme: %s", err))
	}
	k8sCodec = serializer.NewCodecFactory(s)
}

func parseK8sObjects(data []byte) (interface{}, error) {
	decode := k8sCodec.UniversalDeserializer().Decode
	obj, _, err := decode(data, nil, nil)
	return obj, err
}
