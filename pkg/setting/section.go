package setting

type ClientSettingS struct {
	CMDB_API_URL string
}

type ServerSettingS struct {
	ETCD_SERVER_HOST string
	ETCD_SERVER_PORT string
}

func (s *Setting) ReadSection(k string, v interface{}) error {
	err := s.vp.UnmarshalKey(k, v)
	if err != nil {
		return err
	}

	return nil
}
