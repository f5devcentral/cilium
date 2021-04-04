// SPDX-License-Identifier: Apache-2.0
// Copyright 2021 Authors of Cilium

package subscriber

import (
	slim_corev1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/api/core/v1"
)

// ServiceHandler is implemented by event handlers responding to K8s Service
// events.
type ServiceHandler interface {
	OnAdd(*slim_corev1.Service) error
	OnUpdate(oldObj, newObj *slim_corev1.Service) error
	OnDelete(*slim_corev1.Service) error
}

// NewService creates a new subscriber list for ServiceHandlers.
func NewService() *ServiceList {
	return &ServiceList{}
}

// Register registers ServiceHandler as a subscriber for reacting to Service
// objects into the list.
func (l *ServiceList) Register(s ServiceHandler) {
	l.Lock()
	l.subs = append(l.subs, s)
	l.Unlock()
}

// NotifyAdd notifies all the subscribers of an add event to a service.
func (l *ServiceList) NotifyAdd(svc *slim_corev1.Service) []error {
	l.RLock()
	defer l.RUnlock()
	errs := make([]error, 0, len(l.subs))
	for _, s := range l.subs {
		if err := s.OnAdd(svc); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// NotifyUpdate notifies all the subscribers of an update event to a service.
func (l *ServiceList) NotifyUpdate(oldSvc, newSvc *slim_corev1.Service) []error {
	l.RLock()
	defer l.RUnlock()
	errs := make([]error, 0, len(l.subs))
	for _, s := range l.subs {
		if err := s.OnUpdate(oldSvc, newSvc); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// NotifyDelete notifies all the subscribers of an update event to a service.
func (l *ServiceList) NotifyDelete(svc *slim_corev1.Service) []error {
	l.RLock()
	defer l.RUnlock()
	errs := make([]error, 0, len(l.subs))
	for _, s := range l.subs {
		if err := s.OnDelete(svc); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// ServiceList holds the ServiceHandler subscribers that are notified when
// reacting to K8s Service resource / object changes in the K8s watchers.
type ServiceList struct {
	list

	subs []ServiceHandler
}
