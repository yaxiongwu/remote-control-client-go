package engine

import (
	"github.com/stianeikeland/go-rpio/v4"
	"fmt"
	"time"
)

type PiControl struct {
	//CarControl func(speed int, direction int)
	pinNumRunINA int
	pinNumRunINB int
	pinNumDirectionIN1 int
	pinNumDirectionIN2 int
	directionX int
	directionY int
	speed int	
}
/*
type IPiControl interface{
	CarControl func(speed int, direction int)
	}
*/
func  Init(pRunINA int,pRunINB int,pDirectIN1 int,pDirectIN2 int) *PiControl{
	err := rpio.Open()	
	if err != nil{
		fmt.Print(err)
		//return _,err
		}
	
	  return &PiControl{
		  pinNumRunINA:pRunINA,
		  pinNumRunINB:pRunINB,
		  pinNumDirectionIN1:pDirectIN1,
		  pinNumDirectionIN2:pDirectIN2	,
		  directionX:0,
		  directionY:0,
		  speed:0,		  
		  };
	}
	
func (pi *PiControl)  Speed(speed int,direction bool,change<- chan int){
	
	}
/*
 * 速度改变时是只需要调整占空比，但是方向的Y改变时，要改变轮子的方向，此时要重新设置轮子的旋转，速度用以前的
 * */
func (pi *PiControl) DirectionControl(newDirectionX int,newDirectionY int) error{
		pinDirectionIN1 := rpio.Pin(pi.pinNumDirectionIN1)
	    pinDirectionIN2 := rpio.Pin(pi.pinNumDirectionIN2)
	    pinDirectionIN1.Output() 
    	pinDirectionIN2.Output() 
	    pinDirectionIN1.PullUp() 
	    pinDirectionIN2.PullUp()
	    
	     pi.directionX=newDirectionX
	     pi.directionY=newDirectionY
	     
		 fmt.Println(" pi.directionX:", pi.directionX)
		 if(pi.directionX >0){//往左		
			  pinDirectionIN1.High()
		      pinDirectionIN2.Low()			
		      time.Sleep(time.Duration(pi.directionX)*time.Millisecond)
		    }else if(pi.directionX <0){ //往右
			  pinDirectionIN1.Low()
			  pinDirectionIN2.High()
		    time.Sleep(time.Duration(-pi.directionX)*time.Millisecond)
			//  time.Sleep(50*time.Millisecond)
			}						
			pinDirectionIN1.Low()
			pinDirectionIN2.Low()
			return nil
	}
func (pi *PiControl) SpeedControl(newSpeed<- chan int) error{
	
	//err := rpio.Open()	
	//if err != nil{
	//	fmt.Print(err)
	//	return err
	//	}
	//fmt.Println(pi.pinNumRunINA,pi.pinNumRunINB)
		
	pinRunINA := rpio.Pin(pi.pinNumRunINA)
	pinRunINB := rpio.Pin(pi.pinNumRunINB)
	
	pinRunINA.Output() // Output mode
	pinRunINB.Output() 	
	
	pinRunINA.PullUp() // 
	pinRunINB.PullUp() 	
	
	fmt.Println("pi.directionY:",pi.directionY)
		//速度控制
		   go func(){				    
			  for{
				  select{
					 case s:=<-newSpeed:
					    pi.speed=s
					    //pinRunINA.Low()
						//pinRunINB.Low()			
					    fmt.Println("pi.speed=:",pi.speed)
					 //case dy:=<-newDirectionY:
					   // pi.directionY=dy
					    //pinRunINA.Low()
						//pinRunINB.Low()			
					    //fmt.Println("pi.directionY:",pi.directionY)
					 // return
					  case <-time.After(100*time.Millisecond-time.Duration(pi.speed)*time.Millisecond):
						//正转
						//fmt.Println("speed time.After pi.directionY:",pi.directionY)
						//if(pi.directionY == 0){//往前跑		
						 // break
						//}else 
						if(pi.directionY >=0){//往前跑		
						  pinRunINA.High()
						  pinRunINB.Low()						
					    }else if(pi.directionY <0){
						  pinRunINA.Low()
						  pinRunINB.High()			
						}
						time.Sleep(time.Duration(pi.speed)*time.Millisecond)
						//停
						pinRunINA.Low()
						pinRunINB.Low()						
			     }//select
		       }//for
		     }()//go func()
		     
		     /*
		      go func(){				    
			  for{
				  select{
					 case dx:=<-newDirectionX:
					     pi.directionX=dx					  
					    fmt.Println(" pi.directionX:", pi.directionX)
					    if(pi.directionX >0){//往左		
						  pinDirectionIN1.High()
						  pinDirectionIN2.Low()						
						  //time.Sleep(10*time.Duration(pi.directionX)*time.Millisecond)
						//  time.Sleep(50*time.Millisecond)
					    }else if(pi.directionX <0){ //往右
						  pinDirectionIN1.Low()
						  pinDirectionIN2.High()
						  //time.Sleep(10*time.Duration(pi.directionX)*time.Millisecond)
						//  time.Sleep(50*time.Millisecond)
						}						
						pinDirectionIN1.Low()
						pinDirectionIN2.Low()				 		
			     }//select
		       }//for
		     }()//go func()
			*/
	
	
	/*
	if(newDirectionX !=0 || newDirectionY != 0){
		pi.directionX = newDirectionX
		pi.directionY = newDirectionY
		
		if(pi.directionX < 0) {//左转
		//左转
		pinDirectionIN1.High()
		pinDirectionIN2.Low()
		time.Sleep(time.Duration(-pi.directionX)*time.Millisecond)
		//停
		pinDirectionIN1.Low()
		pinDirectionIN2.Low()
		 time.Sleep(time.Duration(15-(-pi.directionX))*time.Millisecond)
		}else if(pi.directionX > 0){//右转
		//右转
			//左转
		pinDirectionIN1.Low()
		pinDirectionIN2.High()
		time.Sleep(time.Duration(-pi.directionX)*time.Millisecond)
		//停
		pinDirectionIN1.Low()
		pinDirectionIN2.Low()
		time.Sleep(time.Duration(15-(-pi.directionX))*time.Millisecond)
		}//else if
	}//if(new
		*/
   /*
	pin.Input()       // Input mode
	res := pin.Read() // Read state from pin (High / Low)
	fmt.Print(res)

	pin.Mode(rpio.Output) // Alternative syntax
	pin.Write(rpio.High)  // Alternative syntax
	pin.PullUp()
	pin.PullDown()
	pin.PullOff()

	pin.Pull(rpio.PullUp)
	rpio.Close()
	* */
	return nil
}
