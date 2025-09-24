package imagegraph

type ImageGraphVersion int

func (v *ImageGraphVersion) Next() ImageGraphVersion {
	*v = *v + 1
	return *v
}
