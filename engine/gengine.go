package engine

import (
	"gengine/builder"
	"gengine/core/errors"
	"github.com/sirupsen/logrus"
	"reflect"
	"sync"
)

type Gengine struct {
}

func NewGengine() *Gengine {
	return &Gengine{}
}

type Stag struct {
	StopTag bool
}

/**
sort execute model

	when b is true it means when there are many rules， if one rule execute error，continue to execute rules after the occur error rule
 */
func (g *Gengine) Execute(rb *builder.RuleBuilder, b bool) error {
	if len(rb.Kc.RuleEntities) == 0 {
		return nil
	}

	for _,r := range rb.Kc.SortRules{
		err := r.Execute()
		if err != nil {
			if b {
				logrus.Errorf("rule: %s executed, error: %+v ",r.RuleName, err)
			} else {
				return errors.Errorf("rule: %s executed, error: %+v ",r.RuleName, err)
			}
		}
	}
	return nil
}

/**
sort execute model

when b is true it means when there are many rules， if one rule execute error，continue to execute rules after the occur error rule;
if stopTag become true,it will not continue to execute

stopTag is a name given by user, and user can use it  to control rules execute behavior in rules, it can improve performance

it used in this scene:
where some high priority rules execute finished, you don't want to execute to the last rules, you can use sTag to control it out of gengine

*/
func (g *Gengine) ExecuteWithStopTag(rb *builder.RuleBuilder, b bool, stopTag string) error {
	rb.Dc.Add(stopTag, false)
	if len(rb.Kc.RuleEntities) == 0 {
		return nil
	}

	for _,r := range rb.Kc.SortRules{
		err := r.Execute()
		if err != nil {
			if b {
				logrus.Errorf("rule: %s executed, error: %+v ",r.RuleName, err)
			} else {
				return errors.Errorf("rule: %s executed, error: %+v ",r.RuleName, err)
			}
		}

		value, err := rb.Dc.Get(stopTag)
		if err!= nil {
			//become sort model
			logrus.Errorf("error: %+v ", err)
		}else {
			if !reflect.ValueOf(value).Bool() {
				break
			}
		}
	}
	return nil
}


/**
sort execute model

when b is true it means when there are many rules， if one rule execute error，continue to execute rules after the occur error rule;
if stopTag become true,it will not continue to execute

sTag is a struct given by user, and user can use it  to control rules execute behavior in rules, it can improve performance

it used in this scene:
where some high priority rules execute finished, you don't want to execute to the last rules, you can use sTag to control it out of gengine
 */
func (g *Gengine) ExecuteWithStopTagDirect(rb *builder.RuleBuilder, b bool, sTag *Stag) error {
	if len(rb.Kc.RuleEntities) == 0 {
		return nil
	}

	for _,r := range rb.Kc.SortRules{
		err := r.Execute()
		if err != nil {
			if b {
				logrus.Errorf("rule: %s executed, error: %+v ",r.RuleName, err)
			} else {
				return errors.Errorf("rule: %s executed, error: %+v ",r.RuleName, err)
			}
		}

		if !sTag.StopTag {
			break
		}
	}
	return nil
}



/*
 concurrent execute model
 in this mode, it will not consider the priority  and not consider err control
 */
func (g *Gengine) ExecuteConcurrent(rb * builder.RuleBuilder){
	if len(rb.Kc.RuleEntities) >= 1 {
		var wg sync.WaitGroup
		wg.Add(len(rb.Kc.RuleEntities))
		for _,r := range rb.Kc.RuleEntities {
			rr := r
			go func() {
				e := rr.Execute()
				if e != nil {
					logrus.Errorf("in rule:%s execute rule err:  %+v", r.RuleName, e)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}


/*
 mix model to execute rules

 in this mode, it will not consider the priority，and it also concurrently to execute rules
 first to execute the most high priority rule，then concurrently to execute last rules without consider the priority
*/
func (g *Gengine) ExecuteMixModel(rb * builder.RuleBuilder){
	rules := rb.Kc.SortRules
	if len(rules) > 0 {
		e := rules[0].Execute()
		if e != nil {
			logrus.Errorf("the most high priority rule: [%s]  exe err:%+v",rules[0].RuleName, e)
		}
	}else{
		return
	}

	if (len(rules) - 1) >= 1 {
		var wg sync.WaitGroup
		wg.Add(len(rules) - 1)
		for _,r := range rules[1:] {
			rr := r
			go func() {
				e := rr.Execute()
				if e != nil {
					logrus.Errorf("in rule:%s execute rule err:  %+v", r.RuleName, e)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

/**
 mix execute model

if stopTag become true,it will not continue to execute
stopTag is a name given by user, and user can use it  to control rules execute behavior in rules, it can improve performance

it used in this scene:
where the first rule execute finished, you don't want to execute to the last rules, you can use sTag to control it in gengine

 */
func (g *Gengine) ExecuteMixModelWithStopTag(rb * builder.RuleBuilder, stopTag string){
	rb.Dc.Add(stopTag, false)
	rules := rb.Kc.SortRules
	if len(rules) > 0 {
		e := rules[0].Execute()
		if e != nil {
			logrus.Errorf("the most high priority rule: [%s]  exe err:%+v",rules[0].RuleName, e)
		}
	}else{
		return
	}

	value, err := rb.Dc.Get(stopTag)
	if err!= nil {
		//become sort model
		logrus.Errorf("error: %+v ", err)
	}else {
		if !reflect.ValueOf(value).Bool() {
			if (len(rules) - 1) >= 1 {
				var wg sync.WaitGroup
				wg.Add(len(rules) - 1)
				for _,r := range rules[1:] {
					rr := r
					go func() {
						e := rr.Execute()
						if e != nil {
							logrus.Errorf("in rule:%s execute rule err:  %+v", r.RuleName, e)
						}
						wg.Done()
					}()
				}
				wg.Wait()
			}
		}
	}
}
/**
 mix execute model

base type :golang translate value
not base type: golang translate pointer

if stopTag become true,it will not continue to execute
stopTag is a name given by user, and user can use it  to control rules execute behavior in rules, it can improve performance

it used in this scene:
where the first rule execute finished, you don't want to execute to the last rules, you can use sTag to control it out of gengine

 */
func (g *Gengine) ExecuteMixModelWithStopTagDirect(rb * builder.RuleBuilder, sTag *Stag){

	rules := rb.Kc.SortRules
	if len(rules) > 0 {
		e := rules[0].Execute()
		if e != nil {
			logrus.Errorf("the most high priority rule: [%s]  exe err:%+v",rules[0].RuleName, e)
		}
	}else{
		return
	}

	if !sTag.StopTag {
		if (len(rules) - 1) >= 1 {
			var wg sync.WaitGroup
			wg.Add(len(rules) - 1)
			for _,r := range rules[1:] {
				rr := r
				go func() {
					e := rr.Execute()
					if e != nil {
						logrus.Errorf("in rule:%s execute rule err:  %+v", r.RuleName, e)
					}
					wg.Done()
				}()
			}
			wg.Wait()
		}
	}
}
