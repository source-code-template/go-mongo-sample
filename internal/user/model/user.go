package model

import "time"

type User struct {
	Id          string     `yaml:"id" mapstructure:"id" json:"id,omitempty" gorm:"column:id;primary_key" bson:"_id,omitempty" dynamodbav:"id" firestore:"-" avro:"id" operator:"="`
	Username    string     `yaml:"username" mapstructure:"username" json:"username,omitempty" gorm:"column:username" bson:"username,omitempty" dynamodbav:"username" firestore:"username" avro:"username" validate:"required,username,max=100"`
	Email       string     `yaml:"email" mapstructure:"email" json:"email,omitempty" gorm:"column:email" bson:"email,omitempty" dynamodbav:"email" firestore:"email" avro:"email" validate:"email,max=100"`
	Phone       string     `yaml:"phone" mapstructure:"phone" json:"phone,omitempty" gorm:"column:phone" bson:"phone,omitempty" dynamodbav:"phone" firestore:"phone" avro:"phone" validate:"required,phone,max=18" operator:"like"`
	DateOfBirth *time.Time `yaml:"date_of_birth" mapstructure:"date_of_birth" json:"dateOfBirth,omitempty" gorm:"column:date_of_birth" bson:"dateOfBirth,omitempty" dynamodbav:"dateOfBirth" firestore:"dateOfBirth" avro:"dateOfBirth"`
}
