package services

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	mock_vision "github.com/manabie-com/backend/mock/golibs/vision"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	vpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestVisionReaderService_DetectTextFromImage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	visionFactory := &mock_vision.VisionFactory{}
	s := &VisionReaderService{
		VisionFactory: visionFactory,
	}
	b64 := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAZsAAAAzCAYAAABBoKXPAAAABHNCSVQICAgIfAhkiAAAABl0RVh0U29mdHdhcmUAZ25vbWUtc2NyZWVuc2hvdO8Dvz4AAA+LSURBVHic7d17XFTlusDxH3s6a2LvRdaQBpKgBZYiJ0QjbyM3MTO18o56vLUrc1unU5onza2G5k6rnbbLayaJN44aeAnzsuOmoGJYgqXgRwZDMHW1kTmbWOczm/PHIA43Q5nZsu35/jlr5l3vu2Z4n/W+7/Mu3KqqqqoQQgghXOg3t7oCQgghbn8SbIQQQricBBshhBAuJ8FGCCGEy0mwEUII4XISbIQQQricBBshhBAuJ8FGCCGEy0mwEUII4XISbIRwAXNEv1tdBSFaFAk2QgghXE6CjRBCCJeTYCOEEMLlbk2wsR3n/WH9MEeM5INj8tBp0cLZzrBqXDTmiIHEZsjvVYib4aRgo3NixVSeGDiSmTsszilSCCHEbcNJweYnck/kY63QOHa8CN05hQohhLhN3OGcYu5j2PR5GA9fon14HxTnFCqEEOI24aRgA0qH3gzt4KzShBBC3E6cFGx09s4ZSmzGz7SfsJz1EwOaV5z1DPs2xLM9PYfTpVZQTHj7dyZs0Fhi+gegXn1f5VH+NOYNdmsmBr+zntdDjXULYu/s0cQe+pn245az/lmHelnPkbo9nu0H7OfQ3U34PRzK4LFjeSbI8xcqeI645yazpqCKtiOXsuXFwDrHdVLmj2ZOyhU8+s0lcbb52mivqW0DeyJFzAw+v/g7Br/zOa+HutU+zY87eXnMUnLcgnlty7s8bfrFKwv6BTL/ZwObD2RxuljDiorJJ4Be4UMYP9qMd0PDUptG7q6NbE7O4tjZ0lqfGT3CjJ97E84LHHp7KDP3XSFoyicsCPqGuM+SycyzUFIBqpc/4YMn8tyoEExYyU1ax2dJaZwo0rAaVLz9g+kf8wLj+3jXGjmnzBvEnNQGvl+wL+xPnsL6on8jfN5uYsPqXL/S42zesJW9h3OxaFao/g08PmoSo7o39hvQ0bJ3snpjMoe+L0GzKXh3eIT+oyczPtxPRvVCNMJpIxunKdrD7OnvkXaxCgwq3j6+KLqG5UQGn53IYN/BV3l39kB8FcD4CNF97mX3jkscSs1BD+1R+4+9PJvUnJ/B0I7+kf7XXi/+ivkz32V/cSUoJvx8vKCslPzDO3g/O4Ovpy8ldoD3dSrZjujIQNYU5HL+UDoFzwfib3A4XHGc1JwrgErPyNBrdbqRtrlCxXfETZ/FmpPlYFAw+fgTYNOwnM1h19kc9maOYdmSyQQ6RryKM2yaPZ2Pc8oBBdXLlwDFiqUoh12f5rD3rwOYu2Q6fVs3vRrn9y9i0uoCNIOKt0lF1TWsxSfZteJNcrXXGPnjOhanlIC7Ce/WJrioUXIyg7g5J7HMWk5s9C/dDPwy67dxzJi9nlwroKh4+/qilJWSn72H/OwMMp9bwuIxAfWCx9mkN5mUnYOmqJhaqahlGiWnDhI3/ztK3JYzJ6z5dRPidtSy9tno37Fqjr0zVkPGsCwhkYS4tcRvSiTp/cmEtIKStPeZ/UledRKCQkiUGW/g8uEM8upkJpQfTuPrCuCBKMI6VN/V2s4QN88eaLwjXiV+WwLxa9cSvy2BT6b1xGTTSPnwPT4vvX6Ka9twM4EKUJLO3tO131v5zUG+LgPufpToYONNts35Crd9aA8090ewMD6RpLgVrI1PIGnFy/Qwgf79Rv60Jd/hE1YOLZttDzRtQnl1xRaSN61lbVwCyfHzGfWQEd2yh/l/jKfI1vR6XC7QeHjyYpK+SCRhUwLJOzcQO6g9CjqFCYtYnF5JrylLSd6ZYD+euJJXenkCGikbkii4gXM1yHqEvyyIJ9eq4Pf4G8QnJpKwdi3x2xJZP6Mf3gYrx9a+zZrv6v4GdPJzLDw8eQEJOxNJSkggedtKXnnMA9DYuzmZ0mZWTYjbVYsKNpe/XMeWoipoE8Hc+ZPp6jAtZOo6hgUznsQEFCZ9yu6L9tfdAiPp6+MGl4+QeqLSoTSdYwePYwU6hpvxrX61fP8GNhdUgt9w5v73QPxq7uJVOg6byR/Md8Hfj7NrX9H1K+sdSXSgEWwXSE0/WetQbuZhNMCzVxQh7jffNufSOX3K3qb2USPp63Xtnl19aAj/NaYbqrvCpVN51zpMy05W77sEhnbEzJvLMw951HxG8erNtLdepq8K+vdbicsob3JN1L4vMGdMCKaro0HlPsKfH0eP6u/CI+Il5owKRL16XH2QYc8OtI8ei09yquymLkCN819sZN/FKnhgOPNmRDlMAyq0H/gqc4a0Bds59ibn1Pts+2Fzif2PHtemG9UHGTZxIH4G4OxJciuaVzchblctKNhYOXTQflffNvzpmo7HkUePITzp6waVJ8k6Ut25GTrRP8wP0EhJdRgV/D2b1OwrcEcn+oddDTU6mRlH7AEo6gn7yKQWlZ6P2uf9C0/kYb1ufe8hPCIEBTifnn7tbtv2HZlHLgMmekV2rZ6Gucm2OZWCp8l+4vPZ6RTUaVzbYe+Q/MUXJC9+Gq/q12raFfAEQzvVXQ8D2kQwNOxewEpmxvEmj8ju8vKi3iX4bTv8TPbRZ1BwcP3jXn7cawAoRytrzsbKnzh2tAAd6BgVVXv6EwCFoMGTGD9iBNH+xjptUugY1Ln+uoyXL94AtnLKr/+jEeJXq+Ws2djOYbFUAgodH/Jv+D2GB+ni7wFFVyg8WwTYF+Y7hpvx21KI5XAGebYQuhqgMieNLCso/24mzMet5hyFRfbRz/kv5jM5rYHFEav9vl7/m4YG9Ts9B57mKHoszyStJJ29Z17Av6Mb5GeQdbEK2oQSHWRsdtucqdvoSfTNfJe0vI1MGvlXujz2GD2DuxLSPZguPvVbWnjWfi08H/avCUC1KXTs7Ae7L1H+g4VSro0gb5hBwWiwl4nS8KKVAlDVzElGWymWH+zfRXvfRmrbIYLnpkY0vUyHgKXbqgC3Rt8qxK9Vywk2VGKtqAIUPNTGV8iN1ccqKxzu/gMiie4Qz5qCDFJPvETX4P/jWMYRrCh0C49y6CjLsVZPc1hLi3BcnahH19Ft1OpI6rm7O2Hd7yIt/QKZKQVM7RjA6fTDWGzg2SuSrjXNaEbbnMn7cRau9mXXhg1sP/A1uSlJ5KYkAQqmh3oTM3kKo0OvLnDrWH+2d+weqkejRSruv0MB9IomXK8WoRJdrwJUPJqYRSeEaL4WFGyMqO5uUKZTqetAA9M26JRb7fMURnfHDrAd0WEBrCk4TUpqDq90spF25AoYg4mulR1krL5pVugf+zlz+jR0jhuhEtYvlD+n76cw/QAFE+8k84gFuI/wqK5OapuTterEoKkLGDRVRzt7kmM5OWQe2EPqya/4aNY35M28mu2loN5ZXbOKxkcTuvV/7VNN7gpKiw80AEYUdzdAp1wedSHEP03LWbMxtMPvfiOOC9n12M5xuqB6CqRD7SmQtlFRBBrg8qEMMrMzyNRA6RpJT8e9JwYv/Noo1z/HDTI+GklYK+xZaSnpZJ6tAm8z0Z0cplJuum1G53XgNh3rRQ3torUmk8/UIZjooZP440fr+WBUe7BppGxO4moNr9bjfH4eWoOF6pw+ZX8Wnsf9fo1MtbnY1evT1Aw1gxd+3kbASklRI7ljFRqWggIsxbIAI4SztJxgg0qv3oEoQOHeJLIa+DuvPLqD/UVVYOxMj9A6d//ekUR3NsKPGaxel4mGQvcwM7X3Od5Dtx4B1edIIK3BrKafyM22ND392D2Y6D73gu0CO1cmkGeDtmYzgbWCxE22zaDi2QrASmFR/e5eLy2itKmd7N++ZFbMSJ6KeYnPztRdYFfo0i3Qfq3KNC5Xl9m2l9m+gJ6XzPZTldRz+RDbMy4BKj37BN+SDY2eJvs3XFJcVD+ho+ICJfWSCe6hW4h93ezEgeQGU7YLE95k3HNTeHFdjjznTwgnaUHBBjwfn8gzvm7w4x4WvRVPrkP/aj2xlTeX7KYEaP/UJJ6st4nwHsIju6CgkV+gwW8fJbpP/emotgPHEO3lBj9+xaLZK8gqduhOys6wbcHL/OeMF3l9a0kTa60QEmkPalbNCob7CIvq7KS2taNLkH1z6YmkFaTV7P3R0XK2Mn9RIk2tJZ59COt6J9jOsW3Zp+Q4pldbz7B9WxYaoDzQ2Z7GC/DAYCZE3Qu2c8TNf5vPTzmsJV08wqr5S0krA+Xh4Uxo4Fr/M3QM6owKVGbF81n25ZrX9eIjxP1xKfsbuKFoO3AE/UxA/lZmLz2ApSZdWcdy4C+8teU0GO6j/5OhtySACnE7akFrNoB7J56PfY2Sme+RdnQdL47cirevF0pFKZZS+32rqfc0Fj4b2GAn4GmOJGT5MbJ08Hisb4MpxqihvBr7EtqsD8nM28qMcTsw+friabBPq1htoAQMZ0L09Z4gUJvbI5FEeyWypbQK7o+i/4MNZCPdZNuCho6lx74lZP3wFbPHHcTU2gRWDc2qo7T2wkRpI1Ncdd3DsFde4+vpb5P27UZejtmKyae63cWlWHXg7mCefzHKYTSoEv7KQqZenM7HOQd5f8pRVvl44W2wUlKs2a/V/RG8MXcsvrdovcbYeyzjO6fz8clzbJoxip2tvfCwWSnRqh8F1ApK6gacVr15bd6zlM76hNydixj35Ur8fFR0rZSSMh1Q6TZlHlODm7umJ4S4qmUFG0DxHcDC1QHVzw/L43RxAToq3p37EDZ4LBMGBDSejuzZm+juH5J16A56hvVp9H2K/xAWrw5k1+Z4dh/Ko7C0gHxUTL5dGdRvGOOH92j4GWGNMXQiLNSTLTsu0d5sbmDvRjPa5v04C5cprF6ZwN5vLWgXNdTWvoT1H0DM6FZsHLeQtH80sZ4+ESxc7cu+zQnsSs8ht7gADQXVy5+wxwYwbOzTtTabAuD+IDFLPiHo6rPRiorI1xVMPp2JDn+KCaOjmvxsNJcwtCPmnWWoa1ayJT0Xi1aKrnrRJXwIw8b2o+zj3/NB/b2ZqEExLF/bie0btrLrcC6WoiJQvehi7sOTI0Yw6BefjyeEuBFuVVVV8q8Hm+0Cm/4wmY9PtWHCyrX8vqGRjfhVMUf0I/2r/be6GkK0GC1qzeZf1vd72H2qEh5oZApNCCF+5VrcNNq/Gr30CKveS8BiU+g2KOrmd9ALIcRtTILNTao6+hGj3tpNibV610rgRKZd998SCCHEr5cEm5tlVO2L+e4muvQdzrRpI/GXPFkhhGiQUxIE/rFqa5Pf+5vnhzf3dEK0eJIgIERtkiAghAtIoBGiNgk2QgghXE722QghhHA5GdkIIYRwOQk2QgghXE6CjRBCCJeTYCOEEMLlJNgIIYRwOQk2QgghXE6CjRBCCJeTYCOEEMLlJNgIIYRwOQk2QgghXO7/AfAN9qZkH8hoAAAAAElFTkSuQmCC"
	testCases := []TestCase{
		{
			name: "err wrong base64 format",
			ctx:  ctx,
			req: &epb.DetectTextFromImageRequest{
				Src:  "wrong base64",
				Lang: "en",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "src must be base64: illegal base64 data at input byte 5"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "err wrong language",
			ctx:  ctx,
			req: &epb.DetectTextFromImageRequest{
				Src:  base64.StdEncoding.EncodeToString([]byte("test")),
				Lang: "wrong language",
			},
			expectedErr: status.Errorf(codes.InvalidArgument, "lang must be en or ja"),
			setup: func(ctx context.Context) {
			},
		},
		{
			name: "err DetectTextFromImage",
			ctx:  ctx,
			req: &epb.DetectTextFromImageRequest{
				Src:  b64,
				Lang: "en",
			},
			expectedErr: status.Errorf(codes.Internal, ErrSomethingWentWrong.Error()),
			setup: func(ctx context.Context) {
				visionFactory.On("DetectTextFromImage", ctx, mock.Anything, mock.Anything).Once().Return(nil, ErrSomethingWentWrong)
			},
		},
		{
			name: "happy case",
			ctx:  ctx,
			req: &epb.DetectTextFromImageRequest{
				Src:  b64,
				Lang: "en",
			},
			setup: func(ctx context.Context) {
				visionFactory.On("DetectTextFromImage", ctx, mock.Anything, mock.Anything).Once().Return(&vpb.TextAnnotation{}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*epb.DetectTextFromImageRequest)
			_, err := s.DetectTextFromImage(testCase.ctx, req)
			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedErr, err)
			}
		})
	}
}
