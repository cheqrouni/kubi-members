package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func DefaultLabelSelector() (selector labels.Selector) {
	selector, _ = metav1.LabelSelectorAsSelector(&metav1.LabelSelector{})
	return
}
