package owl

import (
	"fmt"
	cache "github.com/crosserclaws/intest/common/ccache"
	owlDb "github.com/crosserclaws/intest/common/db/owl"
	owlModel "github.com/crosserclaws/intest/common/model/owl"
)

type GroupTagService struct {
	cache       *cache.CacheCtrl
	cacheConfig *cache.DataCacheConfig
}

func NewGroupTagService(cacheConfig cache.DataCacheConfig) *GroupTagService {
	return &GroupTagService{
		cacheConfig: &cacheConfig,
		cache:       cache.NewCacheCtrl(cache.NewDataCache(cacheConfig)),
	}
}

func (s *GroupTagService) GetGroupTagById(groupTagId int32) *owlModel.GroupTag {
	v := s.cache.MustFetchNativeAndDoNotCacheEmpty(
		groupTagKeyById(groupTagId),
		s.cacheConfig.Duration,
		func() interface{} {
			return owlDb.GetGroupTagById(groupTagId)
		},
	)

	if v == nil {
		return nil
	}

	return v.(*owlModel.GroupTag)
}

func (s *GroupTagService) GetGroupTagsByIds(groupTagIds ...int32) []*owlModel.GroupTag {
	result := make([]*owlModel.GroupTag, 0)

	for _, id := range groupTagIds {
		result = append(result, s.GetGroupTagById(id))
	}

	return result
}

func groupTagKeyById(id int32) string {
	return fmt.Sprintf("!gid!%d", id)
}
