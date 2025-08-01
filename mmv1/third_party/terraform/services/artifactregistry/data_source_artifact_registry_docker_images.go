package artifactregistry

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-google/google/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google/google/transport"
)

func DataSourceArtifactRegistryDockerImages() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceArtifactRegistryDockerImagesRead,
		Schema: map[string]*schema.Schema{
			"location": {
				Type:     schema.TypeString,
				Required: true,
			},
			"repository_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"docker_images": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"self_link": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"image_size_bytes": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"media_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"upload_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"build_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"update_time": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceArtifactRegistryDockerImagesRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*transport_tpg.Config)
	userAgent, err := tpgresource.GenerateUserAgentString(d, config.UserAgent)
	if err != nil {
		return err
	}

	project, err := tpgresource.GetProject(d, config)
	if err != nil {
		return err
	}

	basePath, err := tpgresource.ReplaceVars(d, config, "{{ArtifactRegistryBasePath}}")
	if err != nil {
		return fmt.Errorf("Error setting Artifact Registry base path: %s", err)
	}

	resourcePath, err := tpgresource.ReplaceVars(d, config, fmt.Sprintf("projects/{{project}}/locations/{{location}}/repositories/{{repository_id}}/dockerImages"))
	if err != nil {
		return fmt.Errorf("Error setting resource path: %s", err)
	}

	urlRequest := basePath + resourcePath

	headers := make(http.Header)
	dockerImages := make([]map[string]interface{}, 0)
	pageToken := ""

	for {
		u, err := url.Parse(urlRequest)
		if err != nil {
			return fmt.Errorf("Error parsing URL: %s", err)
		}

		q := u.Query()
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		u.RawQuery = q.Encode()

		res, err := transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
			Config:    config,
			Method:    "GET",
			RawURL:    u.String(),
			UserAgent: userAgent,
			Headers:   headers,
		})

		if err != nil {
			return fmt.Errorf("Error listing Artifact Registry Docker images: %s", err)
		}

		if items, ok := res["dockerImages"].([]interface{}); ok {
			for _, item := range items {
				image := item.(map[string]interface{})

				name, ok := image["name"].(string)
				if !ok {
					return fmt.Errorf("Error getting Artifact Registry Docker image name: %s", err)
				}

				lastComponent := name[strings.LastIndex(name, "/")+1:]
				imageName := strings.SplitN(strings.Split(lastComponent, "@")[0], ":", 2)[0]

				var tags []string
				if rawTags, ok := image["tags"].([]interface{}); ok {
					for _, tag := range rawTags {
						if tagStr, ok := tag.(string); ok {
							tags = append(tags, tagStr)
						}
					}
				}

				getString := func(m map[string]interface{}, key string) string {
					if v, ok := m[key].(string); ok {
						return v
					}
					return ""
				}

				dockerImages = append(dockerImages, map[string]interface{}{
					"image_name":       imageName,
					"name":             name,
					"self_link":        getString(image, "uri"),
					"tags":             tags,
					"image_size_bytes": getString(image, "imageSizeBytes"),
					"media_type":       getString(image, "mediaType"),
					"upload_time":      getString(image, "uploadTime"),
					"build_time":       getString(image, "buildTime"),
					"update_time":      getString(image, "updateTime"),
				})
			}
		}

		if nextToken, ok := res["nextPageToken"].(string); ok && nextToken != "" {
			pageToken = nextToken
		} else {
			break
		}
	}

	if err := d.Set("project", project); err != nil {
		return fmt.Errorf("Error setting project: %s", err)
	}

	if err := d.Set("docker_images", dockerImages); err != nil {
		return fmt.Errorf("Error setting Artifact Registry Docker images: %s", err)
	}

	d.SetId(resourcePath)

	return nil
}
