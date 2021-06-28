// Copyright 2021 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package speaker

import (
	"errors"
	"sync"

	"github.com/cilium/cilium/pkg/cidr"
	"github.com/cilium/cilium/pkg/logging/logfields"

	metallbbgp "go.universe.tf/metallb/pkg/bgp"
	metallbspr "go.universe.tf/metallb/pkg/speaker"
	"golang.org/x/sync/errgroup"
)

var (
	emptyAdverts = []*metallbbgp.Advertisement{}
)

func (s *Speaker) withDraw() error {
	var wg sync.WaitGroup // waitgroup here since we don't care about errors
	for _, session := range s.PeerSessions() {
		wg.Add(1)
		go func(sess metallbspr.Session) { // Need an outer closure to capture session.
			defer wg.Done()
			// providing an empty array or advertisements will
			// provoke the BGP controller to withdrawal any currently
			// advertised bgp routes.
			err := sess.Set(emptyAdverts...)
			if err != nil {
				log.WithError(err).Error("Failed to withdraw BGP routes")
			}
		}(session)
	}
	wg.Wait()
	return nil
}

func (s *Speaker) announcePodCIDRs(cidrs []string) error {
	var eg errgroup.Group
	for _, session := range s.PeerSessions() {
		func(sess metallbspr.Session) { // Need an outer closure to capture session.
			eg.Go(func() error {
				err := s.announce(sess, cidrs)
				if err == nil {
					log.WithField(logfields.CIDR, cidrs).Debug("Announced Pod CIDRs")
				}
				return err
			})
		}(session)
	}

	return eg.Wait()
}

func (s *Speaker) announce(session metallbspr.Session, cidrs []string) error {
	adverts := make([]*metallbbgp.Advertisement, 0, len(cidrs))
	for _, c := range cidrs {
		parsed, err := cidr.ParseCIDR(c)
		if err != nil {
			log.WithError(err).WithField(logfields.CIDR, c).
				Error("Could not announce malformed CIDR")
			continue
		}
		adverts = append(adverts, &metallbbgp.Advertisement{
			Prefix: parsed.IPNet,
		})
	}
	if len(adverts) == 0 {
		return errors.New("no BGP advertisements made")
	}
	return session.Set(adverts...)
}
